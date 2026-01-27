package services

import (
	"fmt"
	"sort"
	"time"

	"incident-teller/internal/domain"
)

// EnhancedTimelineBuilder creates detailed incident timelines with AI insights
type EnhancedTimelineBuilder struct {
	grouper *AlertGrouper
}

// NewEnhancedTimelineBuilder creates a new timeline builder
func NewEnhancedTimelineBuilder(grouper *AlertGrouper) *EnhancedTimelineBuilder {
	return &EnhancedTimelineBuilder{
		grouper: grouper,
	}
}

// TimelineEvent represents an event in the incident timeline
type TimelineEvent struct {
	Timestamp            time.Time
	Type                 string // "trigger", "escalation", "propagation", "resolution", "state_change"
	Severity             string // "info", "warning", "critical"
	Message              string
	SourceAlert          *domain.Alert
	RelatedAlerts        []domain.Alert
	ResourcesAffected    []string
	IsRootCause          bool
	IsCascadePoint       bool
	CausedByEventIndex   *int // Index of the event that caused this one
	TimeFromIncidentStart time.Duration
}

// TimelineWithInsights includes timeline events and AI-generated insights
type TimelineWithInsights struct {
	Events              []TimelineEvent
	StartTime           time.Time
	EndTime             time.Time
	Duration            time.Duration
	CriticalPoints      []int // Indices of critical escalation points
	RootCauseEventIndex *int  // Index of likely root cause event
	ResolutionEventIndex *int // Index of resolution event
}

// BuildTimeline creates a detailed timeline from alerts with AI insights
func (etb *EnhancedTimelineBuilder) BuildTimeline(alerts []domain.Alert, groups []AlertGroup) TimelineWithInsights {
	if len(alerts) == 0 {
		return TimelineWithInsights{
			Events:    []TimelineEvent{},
			StartTime: time.Now(),
			EndTime:   time.Now(),
			Duration:  0,
		}
	}

	// Sort alerts by time
	sortedAlerts := make([]domain.Alert, len(alerts))
	copy(sortedAlerts, alerts)
	sort.Slice(sortedAlerts, func(i, j int) bool {
		return sortedAlerts[i].OccurredAt.Before(sortedAlerts[j].OccurredAt)
	})

	// Build events
	events := etb.buildEvents(sortedAlerts, groups)

	// Identify critical points and root cause
	startTime := sortedAlerts[0].OccurredAt
	criticalPoints := etb.identifyCriticalPoints(events, startTime)
	rootCauseIdx := etb.identifyRootCause(events, groups)
	resolutionIdx := etb.identifyResolution(events)

	return TimelineWithInsights{
		Events:               events,
		StartTime:            startTime,
		EndTime:              sortedAlerts[len(sortedAlerts)-1].OccurredAt,
		Duration:             sortedAlerts[len(sortedAlerts)-1].OccurredAt.Sub(startTime),
		CriticalPoints:       criticalPoints,
		RootCauseEventIndex:  rootCauseIdx,
		ResolutionEventIndex: resolutionIdx,
	}
}

// buildEvents converts alerts to timeline events
func (etb *EnhancedTimelineBuilder) buildEvents(alerts []domain.Alert, groups []AlertGroup) []TimelineEvent {
	events := []TimelineEvent{}
	firstTime := alerts[0].OccurredAt

	for i, alert := range alerts {
		event := TimelineEvent{
			Timestamp:             alert.OccurredAt,
			SourceAlert:           &alert,
			TimeFromIncidentStart: alert.OccurredAt.Sub(firstTime),
			ResourcesAffected:     []string{alert.Host},
		}

		// Determine event type and severity
		if alert.OldStatus == domain.StatusClear && alert.Status != domain.StatusClear {
			event.Type = "trigger"
			event.Severity = string(alert.Status)
		} else if alert.Status == domain.StatusCritical && alert.OldStatus != domain.StatusCritical {
			event.Type = "escalation"
			event.Severity = "critical"
		} else if alert.Status == domain.StatusClear && alert.OldStatus != domain.StatusClear {
			event.Type = "resolution"
			event.Severity = "info"
		} else {
			event.Type = "state_change"
			event.Severity = string(alert.Status)
		}

		// Check for cascade points
		event.IsCascadePoint = etb.isCascadePoint(alert, alerts, i)

		// Generate message
		event.Message = etb.generateEventMessage(alert, event.Type)

		// Find related alerts (within 2 seconds)
		for _, other := range alerts {
			if other.ID != alert.ID && other.OccurredAt.Sub(alert.OccurredAt) > 0 && other.OccurredAt.Sub(alert.OccurredAt) < 2*time.Second {
				event.RelatedAlerts = append(event.RelatedAlerts, other)
			}
		}

		events = append(events, event)
	}

	return events
}

// isCascadePoint checks if this alert triggered other alerts
func (etb *EnhancedTimelineBuilder) isCascadePoint(alert domain.Alert, allAlerts []domain.Alert, index int) bool {
	if index >= len(allAlerts)-1 {
		return false
	}

	nextAlert := allAlerts[index+1]

	// Check if next alert is on same host and within 5 seconds
	if alert.Host != nextAlert.Host {
		return false
	}

	if nextAlert.OccurredAt.Sub(alert.OccurredAt) > 5*time.Second {
		return false
	}

	// Check if it's an escalation (warning to critical)
	if alert.Status == domain.StatusWarning && nextAlert.Status == domain.StatusCritical {
		return true
	}

	// Check if it's a resource cascade
	if etb.isResourceCascade(alert, nextAlert) {
		return true
	}

	return false
}

// isResourceCascade checks if alert is a resource cascade pattern
func (etb *EnhancedTimelineBuilder) isResourceCascade(source, target domain.Alert) bool {
	cascadePatterns := map[domain.ResourceType][]domain.ResourceType{
		domain.ResourceCPU:    {domain.ResourceProcess, domain.ResourceNetwork},
		domain.ResourceMemory: {domain.ResourceProcess, domain.ResourceDisk},
		domain.ResourceDisk:   {domain.ResourceProcess, domain.ResourceNetwork},
	}

	if targets, exists := cascadePatterns[source.ResourceType]; exists {
		for _, t := range targets {
			if target.ResourceType == t {
				return true
			}
		}
	}

	return false
}

// identifyCriticalPoints finds critical escalation points in the timeline
func (etb *EnhancedTimelineBuilder) identifyCriticalPoints(events []TimelineEvent, startTime time.Time) []int {
	criticalPoints := []int{}

	for i, event := range events {
		if event.Type == "escalation" || event.Type == "trigger" {
			if event.Severity == "critical" {
				criticalPoints = append(criticalPoints, i)
			}
		}

		// Also mark cascade points as critical
		if event.IsCascadePoint {
			criticalPoints = append(criticalPoints, i)
		}
	}

	return criticalPoints
}

// identifyRootCause identifies the likely root cause event
func (etb *EnhancedTimelineBuilder) identifyRootCause(events []TimelineEvent, groups []AlertGroup) *int {
	// Find the first critical or trigger event
	for i, event := range events {
		if event.Type == "trigger" || (event.Type == "escalation" && event.Severity == "critical") {
			return &i
		}
	}

	// If no explicit trigger, find the earliest event
	if len(events) > 0 {
		idx := 0
		return &idx
	}

	return nil
}

// identifyResolution finds when the incident was resolved
func (etb *EnhancedTimelineBuilder) identifyResolution(events []TimelineEvent) *int {
	// Find the last resolution event or when all alerts cleared
	for i := len(events) - 1; i >= 0; i-- {
		if events[i].Type == "resolution" {
			return &i
		}
	}

	return nil
}

// generateEventMessage creates a human-readable message for an event
func (etb *EnhancedTimelineBuilder) generateEventMessage(alert domain.Alert, eventType string) string {
	switch eventType {
	case "trigger":
		return fmt.Sprintf(
			"Alert triggered: %s on %s reached %.2f (%s)",
			alert.Name,
			alert.Host,
			alert.Value,
			alert.Status,
		)

	case "escalation":
		return fmt.Sprintf(
			"Incident escalated: %s on %s deteriorated from %s to %s",
			alert.Name,
			alert.Host,
			alert.OldStatus,
			alert.Status,
		)

	case "propagation":
		return fmt.Sprintf(
			"Alert propagated: %s on %s (secondary effect)",
			alert.Name,
			alert.Host,
		)

	case "resolution":
		return fmt.Sprintf(
			"Alert resolved: %s on %s returned to normal",
			alert.Name,
			alert.Host,
		)

	case "state_change":
		return fmt.Sprintf(
			"State change: %s on %s changed from %s to %s",
			alert.Name,
			alert.Host,
			alert.OldStatus,
			alert.Status,
		)

	default:
		return fmt.Sprintf("Event: %s on %s (%s)", alert.Name, alert.Host, alert.Status)
	}
}

// FormatTimeline creates a human-readable timeline string
func (etb *EnhancedTimelineBuilder) FormatTimeline(timeline TimelineWithInsights) string {
	output := fmt.Sprintf("Incident Timeline (Duration: %v)\n", timeline.Duration)
	output += fmt.Sprintf("Start: %s\n", timeline.StartTime.Format("15:04:05"))
	output += fmt.Sprintf("End: %s\n\n", timeline.EndTime.Format("15:04:05"))

	output += "Events:\n"
	for i, event := range timeline.Events {
		relativeTime := event.Timestamp.Sub(timeline.StartTime)
		output += fmt.Sprintf("%d. [%s] (+%v) %s - %s\n",
			i+1,
			event.Type,
			relativeTime,
			event.Message,
			event.Severity,
		)

		if event.IsCascadePoint {
			output += "   └─ CASCADE POINT: Multiple downstream alerts triggered\n"
		}

		if timeline.RootCauseEventIndex != nil && *timeline.RootCauseEventIndex == i {
			output += "   └─ ROOT CAUSE: This is likely where the incident started\n"
		}
	}

	return output
}
