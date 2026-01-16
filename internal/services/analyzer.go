package services

import (
	"fmt"
	"sort"
	"time"

	"incident-teller/internal/domain"
)

// PropagationRule defines how resource issues propagate
type PropagationRule struct {
	From           domain.ResourceType
	To             domain.ResourceType
	MaxTimeWindow  time.Duration
	Description    string
}

// Standard resource propagation patterns based on SRE best practices
var propagationRules = []PropagationRule{
	{
		From:          domain.ResourceMemory,
		To:            domain.ResourceDisk,
		MaxTimeWindow: 5 * time.Minute,
		Description:   "Memory pressure can cause swap/disk thrashing",
	},
	{
		From:          domain.ResourceDisk,
		To:            domain.ResourceCPU,
		MaxTimeWindow: 5 * time.Minute,
		Description:   "Disk saturation causes CPU iowait",
	},
	{
		From:          domain.ResourceMemory,
		To:            domain.ResourceCPU,
		MaxTimeWindow: 10 * time.Minute,
		Description:   "Memory pressure can indirectly cause CPU issues",
	},
	{
		From:          domain.ResourceNetwork,
		To:            domain.ResourceMemory,
		MaxTimeWindow: 3 * time.Minute,
		Description:   "Network buffer exhaustion affects memory",
	},
	{
		From:          domain.ResourceProcess,
		To:            domain.ResourceMemory,
		MaxTimeWindow: 2 * time.Minute,
		Description:   "Process leak causes memory pressure",
	},
	{
		From:          domain.ResourceProcess,
		To:            domain.ResourceCPU,
		MaxTimeWindow: 2 * time.Minute,
		Description:   "Runaway process consumes CPU",
	},
}

// IncidentAnalyzer provides SRE-grade incident analysis
type IncidentAnalyzer struct {
	propagationRules []PropagationRule
}

// NewIncidentAnalyzer creates a new analyzer instance
func NewIncidentAnalyzer() *IncidentAnalyzer {
	return &IncidentAnalyzer{
		propagationRules: propagationRules,
	}
}

// AnalyzeIncident takes a list of alerts and produces an ordered timeline with causality
func (a *IncidentAnalyzer) AnalyzeIncident(alerts []domain.Alert) []domain.TimelineEntry {
	if len(alerts) == 0 {
		return []domain.TimelineEntry{}
	}

	// Sort alerts chronologically
	sortedAlerts := make([]domain.Alert, len(alerts))
	copy(sortedAlerts, alerts)
	sort.Slice(sortedAlerts, func(i, j int) bool {
		return sortedAlerts[i].OccurredAt.Before(sortedAlerts[j].OccurredAt)
	})

	// Track active resource issues (resource type -> alert that triggered it)
	activeIssues := make(map[domain.ResourceType]*domain.Alert)

	// Build timeline
	timeline := make([]domain.TimelineEntry, 0, len(sortedAlerts))
	incidentStart := sortedAlerts[0].OccurredAt

	for i := range sortedAlerts {
		alert := &sortedAlerts[i]
		entry := a.createTimelineEntry(alert, incidentStart, activeIssues)
		timeline = append(timeline, entry)

		// Update active issues tracking
		a.updateActiveIssues(activeIssues, alert)
	}

	return timeline
}

// createTimelineEntry generates a timeline entry with causality detection
func (a *IncidentAnalyzer) createTimelineEntry(
	alert *domain.Alert,
	incidentStart time.Time,
	activeIssues map[domain.ResourceType]*domain.Alert,
) domain.TimelineEntry {
	// Calculate duration since incident start
	duration := alert.OccurredAt.Sub(incidentStart)

	// Detect potential causes
	causes := a.detectCauses(alert, activeIssues)

	// Build entry
	entry := domain.TimelineEntry{
		Timestamp:          alert.OccurredAt,
		Type:               determineEventType(alert),
		Message:            formatMessage(alert, causes),
		Severity:           mapSeverity(alert.Status),
		DurationSinceStart: &duration,
		CausedBy:           extractAlertIDs(causes),
		RelatedAlertIDs:    []string{alert.ID},
		ResourceType:       alert.ResourceType,
	}

	return entry
}

// detectCauses finds earlier alerts that likely caused this one
func (a *IncidentAnalyzer) detectCauses(
	alert *domain.Alert,
	activeIssues map[domain.ResourceType]*domain.Alert,
) []*domain.Alert {
	var causes []*domain.Alert

	// Check each propagation rule
	for _, rule := range a.propagationRules {
		// Only check if this alert matches the "To" resource type
		if rule.To != alert.ResourceType {
			continue
		}

		// Check if there's an active issue of the "From" type
		if sourceAlert, exists := activeIssues[rule.From]; exists {
			// Verify time window
			timeSince := alert.OccurredAt.Sub(sourceAlert.OccurredAt)
			if timeSince >= 0 && timeSince <= rule.MaxTimeWindow {
				// Verify the source is still in a problem state
				if sourceAlert.Status != domain.StatusClear {
					causes = append(causes, sourceAlert)
				}
			}
		}
	}

	return causes
}

// updateActiveIssues maintains the state of ongoing resource issues
func (a *IncidentAnalyzer) updateActiveIssues(
	activeIssues map[domain.ResourceType]*domain.Alert,
	alert *domain.Alert,
) {
	if alert.Status == domain.StatusClear {
		// Remove from active issues when cleared
		delete(activeIssues, alert.ResourceType)
	} else {
		// Add or update active issue
		activeIssues[alert.ResourceType] = alert
	}
}

// determineEventType categorizes the alert transition
func determineEventType(alert *domain.Alert) string {
	switch {
	case alert.OldStatus == domain.StatusClear && alert.Status != domain.StatusClear:
		return "TRIGGERED"
	case alert.OldStatus == domain.StatusWarning && alert.Status == domain.StatusCritical:
		return "ESCALATED"
	case alert.Status == domain.StatusClear:
		return "RESOLVED"
	case alert.Status == domain.StatusWarning:
		return "WARNING"
	case alert.Status == domain.StatusCritical:
		return "CRITICAL"
	default:
		return "UPDATE"
	}
}

// formatMessage creates a human-readable message with causality context
func formatMessage(alert *domain.Alert, causes []*domain.Alert) string {
	baseMsg := fmt.Sprintf("[%s] %s on %s (value: %.2f)",
		alert.ResourceType, alert.Name, alert.Chart, alert.Value)

	if alert.Host != "" {
		baseMsg = fmt.Sprintf("[%s@%s] %s on %s (value: %.2f)",
			alert.ResourceType, alert.Host, alert.Name, alert.Chart, alert.Value)
	}

	if len(causes) > 0 {
		// Add causality information
		causeDesc := "\n  ↳ Likely caused by: "
		for i, cause := range causes {
			if i > 0 {
				causeDesc += ", "
			}
			causeDesc += fmt.Sprintf("%s issue (%.1fs earlier)",
				cause.ResourceType,
				alert.OccurredAt.Sub(cause.OccurredAt).Seconds())
		}
		baseMsg += causeDesc
	}

	if alert.Description != "" {
		baseMsg += fmt.Sprintf("\n  Info: %s", alert.Description)
	}

	return baseMsg
}

// mapSeverity converts alert status to severity level
func mapSeverity(status domain.AlertStatus) string {
	switch status {
	case domain.StatusCritical:
		return "critical"
	case domain.StatusWarning:
		return "warning"
	case domain.StatusClear:
		return "success"
	default:
		return "info"
	}
}

// extractAlertIDs extracts IDs from a list of alerts
func extractAlertIDs(alerts []*domain.Alert) []string {
	ids := make([]string, 0, len(alerts))
	for _, alert := range alerts {
		ids = append(ids, alert.ID)
	}
	return ids
}

// GenerateIncidentSummary creates a summary of the incident with root cause analysis
func (a *IncidentAnalyzer) GenerateIncidentSummary(timeline []domain.TimelineEntry) string {
	if len(timeline) == 0 {
		return "No events in timeline"
	}

	// Find root cause (earliest event with no causes)
	var rootCause *domain.TimelineEntry
	for i := range timeline {
		if len(timeline[i].CausedBy) == 0 && timeline[i].Type == "TRIGGERED" {
			rootCause = &timeline[i]
			break
		}
	}

	summary := "=== Incident Timeline Summary ===\n\n"

	if rootCause != nil {
		summary += fmt.Sprintf("🔴 Root Cause: %s at %s\n",
			rootCause.ResourceType,
			rootCause.Timestamp.Format(time.RFC3339))
		summary += fmt.Sprintf("   %s\n\n", rootCause.Message)
	}

	summary += fmt.Sprintf("📊 Total Events: %d\n", len(timeline))
	if len(timeline) > 0 {
		duration := timeline[len(timeline)-1].Timestamp.Sub(timeline[0].Timestamp)
		summary += fmt.Sprintf("⏱️  Duration: %s\n\n", duration.Round(time.Second))
	}

	summary += "📋 Event Sequence:\n"
	for i, entry := range timeline {
		summary += fmt.Sprintf("%d. [%s] %s - %s\n",
			i+1,
			entry.Timestamp.Format("15:04:05"),
			entry.Type,
			entry.Message)
	}

	return summary
}
