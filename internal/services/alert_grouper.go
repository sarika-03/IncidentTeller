package services

import (
	"sort"
	"time"

	"incident-teller/internal/domain"
)

// AlertGrouper groups related alerts based on various criteria
type AlertGrouper struct {
	correlationWindow time.Duration
}

// NewAlertGrouper creates a new alert grouper
func NewAlertGrouper(correlationWindow time.Duration) *AlertGrouper {
	return &AlertGrouper{
		correlationWindow: correlationWindow,
	}
}

// AlertGroup represents a group of related alerts
type AlertGroup struct {
	ID               string
	Alerts           []domain.Alert
	PrimaryHost      string
	AffectedHosts    []string
	ResourceTypes    []domain.ResourceType
	StartTime        time.Time
	EndTime          time.Time
	IsCascading      bool
	CascadeChain     []AlertCascade
	GroupType        string // "single_host", "multi_host", "cascading", "distributed"
}

// AlertCascade represents a cascade relationship between alerts
type AlertCascade struct {
	SourceAlert  domain.Alert
	TargetAlert  domain.Alert
	DelaySeconds float64
	Confidence   float64
	Type         string // "propagation", "dependency", "timeout"
}

// GroupAlerts groups alerts by host, time window, and cascade relationships
func (ag *AlertGrouper) GroupAlerts(alerts []domain.Alert) []AlertGroup {
	if len(alerts) == 0 {
		return []AlertGroup{}
	}

	// Sort alerts by time
	sortedAlerts := make([]domain.Alert, len(alerts))
	copy(sortedAlerts, alerts)
	sort.Slice(sortedAlerts, func(i, j int) bool {
		return sortedAlerts[i].OccurredAt.Before(sortedAlerts[j].OccurredAt)
	})

	// Group by host first
	hostGroups := ag.groupByHost(sortedAlerts)

	// Detect cascading relationships within and across hosts
	groups := ag.detectCascades(hostGroups, sortedAlerts)

	return groups
}

// groupByHost groups alerts by hostname and time window
func (ag *AlertGrouper) groupByHost(alerts []domain.Alert) map[string][]domain.Alert {
	hostGroups := make(map[string][]domain.Alert)

	for _, alert := range alerts {
		hostGroups[alert.Host] = append(hostGroups[alert.Host], alert)
	}

	return hostGroups
}

// detectCascades detects cascade relationships between alerts
func (ag *AlertGrouper) detectCascades(hostGroups map[string][]domain.Alert, allAlerts []domain.Alert) []AlertGroup {
	groups := []AlertGroup{}
	processed := make(map[int]bool)

	for i, alert := range allAlerts {
		if processed[i] {
			continue
		}

		// Start a new group with this alert
		group := AlertGroup{
			ID:           alert.ID,
			Alerts:       []domain.Alert{alert},
			PrimaryHost:  alert.Host,
			AffectedHosts: []string{alert.Host},
			ResourceTypes: []domain.ResourceType{alert.ResourceType},
			StartTime:    alert.OccurredAt,
			EndTime:      alert.OccurredAt,
		}

		processed[i] = true

		// Find related alerts within the correlation window
		for j := i + 1; j < len(allAlerts); j++ {
			if processed[j] {
				continue
			}

			nextAlert := allAlerts[j]

			// Check if within time window
			if nextAlert.OccurredAt.Sub(alert.OccurredAt) > ag.correlationWindow {
				break // Alerts are sorted by time, so we can break
			}

			// Check if it's related (same host, cascading, or resource dependency)
			if ag.isRelated(alert, nextAlert, allAlerts) {
				group.Alerts = append(group.Alerts, nextAlert)
				group.EndTime = nextAlert.OccurredAt

				// Update affected hosts
				if !contains(group.AffectedHosts, nextAlert.Host) {
					group.AffectedHosts = append(group.AffectedHosts, nextAlert.Host)
				}

				// Update resource types
				if !containsResourceType(group.ResourceTypes, nextAlert.ResourceType) {
					group.ResourceTypes = append(group.ResourceTypes, nextAlert.ResourceType)
				}

				// Check for cascade
				cascade := ag.detectCascadeRelationship(alert, nextAlert)
				if cascade != nil {
					group.IsCascading = true
					group.CascadeChain = append(group.CascadeChain, *cascade)
				}

				processed[j] = true
			}
		}

		// Determine group type
		group.GroupType = ag.determineGroupType(group)

		groups = append(groups, group)
	}

	return groups
}

// isRelated checks if two alerts are related
func (ag *AlertGrouper) isRelated(alert1, alert2 domain.Alert, allAlerts []domain.Alert) bool {
	// Same host
	if alert1.Host == alert2.Host {
		return true
	}

	// Cascading relationship
	if ag.isCascading(alert1, alert2) {
		return true
	}

	// Resource dependency (e.g., CPU spike leading to process issues)
	if ag.hasResourceDependency(alert1, alert2) {
		return true
	}

	return false
}

// isCascading checks if alert2 is likely caused by alert1
func (ag *AlertGrouper) isCascading(source, target domain.Alert) bool {
	// Must be on same host or dependent hosts
	if source.Host != target.Host {
		// Could still be cascading if it's a distributed system
		// For now, we require same host for cascading detection
		return false
	}

	// Check time proximity (cascade should happen within a few seconds)
	delay := target.OccurredAt.Sub(source.OccurredAt).Seconds()
	if delay < 0 || delay > 30 { // Within 30 seconds
		return false
	}

	// Check if target severity is higher (escalation pattern)
	if target.Status == domain.StatusCritical && source.Status == domain.StatusWarning {
		return true
	}

	// Check if it's a known cascade pattern (e.g., CPU -> Process)
	if source.ResourceType == domain.ResourceCPU && target.ResourceType == domain.ResourceProcess {
		return true
	}

	if source.ResourceType == domain.ResourceMemory && target.ResourceType == domain.ResourceProcess {
		return true
	}

	if source.ResourceType == domain.ResourceDisk && target.ResourceType == domain.ResourceProcess {
		return true
	}

	return false
}

// hasResourceDependency checks if there's a resource dependency
func (ag *AlertGrouper) hasResourceDependency(alert1, alert2 domain.Alert) bool {
	// CPU/Memory pressure can affect other resources
	if (alert1.ResourceType == domain.ResourceCPU || alert1.ResourceType == domain.ResourceMemory) &&
		(alert2.ResourceType == domain.ResourceNetwork || alert2.ResourceType == domain.ResourceDisk) {
		return true
	}

	return false
}

// detectCascadeRelationship creates a cascade relationship if detected
func (ag *AlertGrouper) detectCascadeRelationship(source, target domain.Alert) *AlertCascade {
	if !ag.isCascading(source, target) {
		return nil
	}

	delay := target.OccurredAt.Sub(source.OccurredAt).Seconds()

	cascadeType := "propagation"
	confidence := 0.7

	// Increase confidence for known patterns
	if (source.ResourceType == domain.ResourceCPU && target.ResourceType == domain.ResourceProcess) ||
		(source.ResourceType == domain.ResourceMemory && target.ResourceType == domain.ResourceProcess) {
		confidence = 0.9
		cascadeType = "dependency"
	}

	return &AlertCascade{
		SourceAlert:  source,
		TargetAlert:  target,
		DelaySeconds: delay,
		Confidence:   confidence,
		Type:         cascadeType,
	}
}

// determineGroupType determines the type of group
func (ag *AlertGrouper) determineGroupType(group AlertGroup) string {
	if len(group.AffectedHosts) > 1 {
		if group.IsCascading {
			return "cascading"
		}
		return "multi_host"
	}

	if group.IsCascading {
		return "cascading"
	}

	if len(group.ResourceTypes) > 1 {
		return "cascading"
	}

	return "single_host"
}

// Helper functions

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func containsResourceType(slice []domain.ResourceType, item domain.ResourceType) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
