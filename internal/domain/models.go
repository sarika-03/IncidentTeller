package domain

import (
	"time"
)

// AlertStatus represents the state of an alert (CLEAR, WARNING, CRITICAL)
type AlertStatus string

const (
	StatusUndefined AlertStatus = "UNDEFINED"
	StatusClear     AlertStatus = "CLEAR"
	StatusWarning   AlertStatus = "WARNING"
	StatusCritical  AlertStatus = "CRITICAL"
	StatusRemoved   AlertStatus = "REMOVED"
)

// ResourceType represents the category of system resource being monitored
type ResourceType string

const (
	ResourceUnknown ResourceType = "UNKNOWN"
	ResourceCPU     ResourceType = "CPU"
	ResourceMemory  ResourceType = "MEMORY"
	ResourceDisk    ResourceType = "DISK"
	ResourceNetwork ResourceType = "NETWORK"
	ResourceProcess ResourceType = "PROCESS"
)

// Alert represents a normalized event ingested from an external source (Netdata)
type Alert struct {
	ID           string       // Unique Event ID
	ExternalID   uint64       // Netdata's unique_id
	Host         string       // Hostname/node where alert originated
	Chart        string       // e.g., "system.cpu"
	Family       string       // e.g., "cpu"
	Name         string       // e.g., "10min_cpu_usage"
	Status       AlertStatus  // Current status
	OldStatus    AlertStatus  // Previous status
	Value        float64      // The metric value triggering the alert
	OccurredAt   time.Time    // When the event happened
	Description  string       // Raw description if available
	ResourceType ResourceType // Classified resource type
	Labels       map[string]string
}

// Incident represents a grouped collection of alerts related to a specific issue
type Incident struct {
	ID         string
	Title      string      // e.g., "High CPU usage on system.cpu"
	Status     AlertStatus // Current aggregate status
	StartedAt  time.Time
	ResolvedAt *time.Time // Nil if active
	Events     []Alert    // Ordered list of events in this incident
}

// TimelineEntry is a human-readable representation of an event in the timeline
type TimelineEntry struct {
	Timestamp          time.Time
	Type               string         // e.g., "TRIGGERED", "ESCALATED", "RESOLVED", "NOTE"
	Message            string         // Human-readable description
	Severity           string         // "info", "warning", "critical", "success"
	DurationSinceStart *time.Duration // Optional: How long into the incident this happened
	CausedBy           []string       // IDs of alerts that likely caused this event
	RelatedAlertIDs    []string       // All related alert IDs for this entry
	ResourceType       ResourceType   // Resource affected
}

// ParsedNetdataResponse represents the raw JSON structure from Netdata (for reference in adapters)
// Placed here for model clarity, usually lives in adapters/netdata but helpful to visualize mapping.
type NetdataAlarmLog struct {
	UniqueID    uint64  `json:"unique_id"`
	AlarmID     uint64  `json:"alarm_id"`
	EventID     uint64  `json:"event_id"`
	When        uint64  `json:"when"` // Unix timestamp
	Name        string  `json:"name"`
	Chart       string  `json:"chart"`
	Family      string  `json:"family"`
	Status      string  `json:"status"`
	OldStatus   string  `json:"old_status"`
	Value       float64 `json:"value"`
	OldValue    float64 `json:"old_value"`
	Updated     bool    `json:"updated"`
	Exec        string  `json:"exec"`
	Recipient   string  `json:"recipient"`
	Source      string  `json:"source"`
	Units       string  `json:"units"`
	Info        string  `json:"info"`
	ValueString string  `json:"value_string"`
	Hostname    string  `json:"hostname"` // Optional, might be in different API versions
}

// NetdataAlarmLogResponse wraps the API response
type NetdataAlarmLogResponse struct {
	Alarms              []NetdataAlarmLog `json:"alarms"`
	LatestAlarmUniqueID uint64            `json:"latest_alarm_log_unique_id"`
}
