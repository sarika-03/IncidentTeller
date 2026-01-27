package repository

import (
	"context"
	"fmt"
	"sync"

	"incident-teller/internal/domain"
)

// InMemoryRepository provides a simple in-memory storage for testing and development
type InMemoryRepository struct {
	mu              sync.RWMutex
	alerts          map[string]domain.Alert // alertID -> Alert
	incidents       []domain.Incident
	lastProcessedID uint64
}

// NewInMemoryRepository creates a new in-memory repository
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		alerts:          make(map[string]domain.Alert),
		incidents:       make([]domain.Incident, 0),
		lastProcessedID: 0,
	}
}

// SaveAlert stores an alert in memory
func (r *InMemoryRepository) SaveAlert(ctx context.Context, alert domain.Alert) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.alerts[alert.ID] = alert
	return nil
}

// GetIncidents returns all stored incidents
func (r *InMemoryRepository) GetIncidents(ctx context.Context) ([]domain.Incident, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return copy to prevent external modification
	incidents := make([]domain.Incident, len(r.incidents))
	copy(incidents, r.incidents)
	return incidents, nil
}

// SaveIncident stores an incident
func (r *InMemoryRepository) SaveIncident(ctx context.Context, incident domain.Incident) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if incident already exists
	for i, existing := range r.incidents {
		if existing.ID == incident.ID {
			r.incidents[i] = incident
			return nil
		}
	}

	// Add new incident
	r.incidents = append(r.incidents, incident)
	return nil
}

// GetLastProcessedID returns the last processed alert ID
func (r *InMemoryRepository) GetLastProcessedID(ctx context.Context) (uint64, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.lastProcessedID, nil
}

// SetLastProcessedID updates the last processed alert ID
func (r *InMemoryRepository) SetLastProcessedID(ctx context.Context, id uint64) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lastProcessedID = id
	return nil
}

// GetAlerts returns all stored alerts (useful for analysis)
func (r *InMemoryRepository) GetAlerts(ctx context.Context) ([]domain.Alert, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	alerts := make([]domain.Alert, 0, len(r.alerts))
	for _, alert := range r.alerts {
		alerts = append(alerts, alert)
	}
	return alerts, nil
}

// GetAlertByID retrieves a specific alert
func (r *InMemoryRepository) GetAlertByID(ctx context.Context, id string) (domain.Alert, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	alert, exists := r.alerts[id]
	if !exists {
		return domain.Alert{}, fmt.Errorf("alert not found: %s", id)
	}
	return alert, nil
}

// Clear removes all data (useful for testing)
func (r *InMemoryRepository) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.alerts = make(map[string]domain.Alert)
	r.incidents = make([]domain.Incident, 0)
	r.lastProcessedID = 0
}

// Stats returns repository statistics
func (r *InMemoryRepository) Stats(ctx context.Context) (map[string]interface{}, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return map[string]interface{}{
		"total_alerts":      len(r.alerts),
		"total_incidents":   len(r.incidents),
		"last_processed_id": r.lastProcessedID,
	}, nil
}

// PingContext checks repository connectivity
func (r *InMemoryRepository) PingContext(ctx context.Context) error {
	return nil // In-memory repo is always available
}
