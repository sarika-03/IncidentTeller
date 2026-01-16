package netdata

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"incident-teller/internal/domain"
)

// Client implements the AlertSource interface for Netdata
type Client struct {
	baseURL    string
	httpClient *http.Client
	hostname   string // Default hostname if not in response
}

// NewClient creates a new Netdata API client
func NewClient(baseURL, hostname string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		hostname: hostname,
	}
}

// FetchLatest retrieves alarm logs from Netdata API since the given unique ID
func (c *Client) FetchLatest(ctx context.Context, lastID uint64) ([]domain.Alert, error) {
	// Build URL with query parameters
	apiURL, err := url.Parse(c.baseURL + "/api/v1/alarm_log")
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}

	query := apiURL.Query()
	if lastID > 0 {
		query.Set("after", fmt.Sprintf("%d", lastID))
	}
	apiURL.RawQuery = query.Encode()

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch alarm log: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Try to parse as array first (common format)
	var logs []domain.NetdataAlarmLog
	if err := json.Unmarshal(body, &logs); err != nil {
		// If that fails, try wrapped response
		var wrappedResp domain.NetdataAlarmLogResponse
		if err := json.Unmarshal(body, &wrappedResp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
		logs = wrappedResp.Alarms
	}

	// Normalize to domain alerts
	alerts := make([]domain.Alert, 0, len(logs))
	for _, log := range logs {
		alert := c.normalizeAlert(log)
		alerts = append(alerts, alert)
	}

	return alerts, nil
}

// normalizeAlert converts a Netdata alarm log entry to domain Alert
func (c *Client) normalizeAlert(log domain.NetdataAlarmLog) domain.Alert {
	// Determine hostname
	hostname := log.Hostname
	if hostname == "" {
		hostname = c.hostname
	}

	// Parse timestamp
	occurredAt := time.Unix(int64(log.When), 0)

	// Map status strings
	status := mapStatus(log.Status)
	oldStatus := mapStatus(log.OldStatus)

	// Classify resource type
	resourceType := classifyResourceType(log.Chart, log.Family)

	// Generate unique ID
	alertID := fmt.Sprintf("%s-%d", hostname, log.UniqueID)

	return domain.Alert{
		ID:           alertID,
		ExternalID:   log.UniqueID,
		Host:         hostname,
		Chart:        log.Chart,
		Family:       log.Family,
		Name:         log.Name,
		Status:       status,
		OldStatus:    oldStatus,
		Value:        log.Value,
		OccurredAt:   occurredAt,
		Description:  log.Info,
		ResourceType: resourceType,
		Labels: map[string]string{
			"source":      log.Source,
			"units":       log.Units,
			"exec":        log.Exec,
			"recipient":   log.Recipient,
			"alarm_id":    fmt.Sprintf("%d", log.AlarmID),
			"event_id":    fmt.Sprintf("%d", log.EventID),
		},
	}
}

// mapStatus converts Netdata status string to domain AlertStatus
func mapStatus(status string) domain.AlertStatus {
	switch status {
	case "CLEAR":
		return domain.StatusClear
	case "WARNING":
		return domain.StatusWarning
	case "CRITICAL":
		return domain.StatusCritical
	case "UNDEFINED":
		return domain.StatusUndefined
	default:
		return domain.StatusUndefined
	}
}

// classifyResourceType determines resource type from chart/family information
func classifyResourceType(chart, family string) domain.ResourceType {
	// Family-based classification
	switch family {
	case "cpu", "cpufreq":
		return domain.ResourceCPU
	case "mem", "ram", "swap":
		return domain.ResourceMemory
	case "disk", "disk_space", "disk_ops", "disk_util", "disk_iotime":
		return domain.ResourceDisk
	case "net", "network", "ipv4", "ipv6":
		return domain.ResourceNetwork
	case "apps", "processes":
		return domain.ResourceProcess
	}

	// Chart-based fallback classification
	switch {
	case contains(chart, "cpu"):
		return domain.ResourceCPU
	case contains(chart, "mem") || contains(chart, "ram") || contains(chart, "swap"):
		return domain.ResourceMemory
	case contains(chart, "disk"):
		return domain.ResourceDisk
	case contains(chart, "net") || contains(chart, "network"):
		return domain.ResourceNetwork
	}

	return domain.ResourceUnknown
}

// contains is a helper to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		(s == substr || (len(s) > len(substr) && 
		(stringContains(s, substr))))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
