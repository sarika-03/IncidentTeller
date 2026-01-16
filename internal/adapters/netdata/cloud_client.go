package netdata

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"incident-teller/internal/domain"
)

// CloudClient implements Netdata Cloud API
type CloudClient struct {
	token      string
	space      string
	rooms      []string
	httpClient *http.Client
	baseURL    string
}

// NewCloudClient creates a new Netdata Cloud client
func NewCloudClient(token, space string, rooms ...string) *CloudClient {
	return &CloudClient{
		token:   token,
		space:   space,
		rooms:   rooms,
		baseURL: "https://app.netdata.cloud/api/v2",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// CloudAlarmResponse represents Netdata Cloud alarm response
type CloudAlarmResponse struct {
	Data struct {
		Alarms []CloudAlarm `json:"alarms"`
	} `json:"data"`
}

type CloudAlarm struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Node      string  `json:"node"`
	Chart     string  `json:"chart"`
	Status    string  `json:"status"`
	OldStatus string  `json:"oldStatus"`
	Value     float64 `json:"value"`
	Timestamp int64   `json:"when"`
	Info      string  `json:"info"`
	Component string  `json:"component"`
	Room      string  `json:"room"`
}

// CloudGraphQLResponse represents GraphQL response structure
type CloudGraphQLResponse struct {
	Data struct {
		Space struct {
			Alarms struct {
				Edges []struct {
					Node CloudAlarm `json:"node"`
				} `json:"edges"`
				pageInfo struct {
					EndCursor   string `json:"endCursor"`
					HasNextPage bool   `json:"hasNextPage"`
				} `json:"pageInfo"`
			} `json:"alarms"`
		} `json:"space"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// FetchLatest retrieves alarms from Netdata Cloud
func (c *CloudClient) FetchLatest(ctx context.Context, lastID uint64) ([]domain.Alert, error) {
	// Build GraphQL query for Cloud API
	query := `
	query GetAlarms($space: String!, $after: String, $rooms: [String!]) {
		space(id: $space) {
			alarms(after: $after, first: 100, rooms: $rooms) {
				edges {
					node {
						id
						name
						node
						chart
						status
						oldStatus
						value
						timestamp
						info
						component
						room
					}
				}
				pageInfo {
					endCursor
					hasNextPage
				}
			}
		}
	}
	`

	variables := map[string]interface{}{
		"space": c.space,
		"after": fmt.Sprintf("%d", lastID),
		"rooms": c.rooms,
	}

	reqBody, err := json.Marshal(map[string]interface{}{
		"query":     query,
		"variables": variables,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create request with authentication
	req, err := http.NewRequestWithContext(ctx, "POST",
		c.baseURL+"/graphql", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("cloud API error %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var cloudResp CloudGraphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&cloudResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for GraphQL errors
	if len(cloudResp.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL errors: %v", cloudResp.Errors)
	}

	// Convert to domain alerts
	alerts := make([]domain.Alert, 0, len(cloudResp.Data.Space.Alarms.Edges))
	for _, edge := range cloudResp.Data.Space.Alarms.Edges {
		alerts = append(alerts, c.normalizeCloudAlarm(edge.Node))
	}

	return alerts, nil
}

// normalizeCloudAlarm converts Cloud alarm to domain alert
func (c *CloudClient) normalizeCloudAlarm(alarm CloudAlarm) domain.Alert {
	status := mapStatus(alarm.Status)
	oldStatus := mapStatus(alarm.OldStatus)
	resourceType := classifyResourceType(alarm.Chart, alarm.Component)

	return domain.Alert{
		ID:           alarm.ID,
		ExternalID:   uint64(alarm.Timestamp), // Use timestamp as unique ID
		Host:         alarm.Node,
		Chart:        alarm.Chart,
		Family:       alarm.Component,
		Name:         alarm.Name,
		Status:       status,
		OldStatus:    oldStatus,
		Value:        alarm.Value,
		OccurredAt:   time.Unix(alarm.Timestamp, 0),
		Description:  alarm.Info,
		ResourceType: resourceType,
		Labels: map[string]string{
			"source":   "netdata-cloud",
			"space":    c.space,
			"room":     alarm.Room,
			"node_id":  alarm.Node,
			"chart_id": alarm.Chart,
		},
	}
}
