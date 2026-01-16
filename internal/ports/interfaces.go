package ports

import (
	"context"
	"incident-teller/internal/domain"
)

// AlertSource defines how we fetch alerts from external systems
type AlertSource interface {
	// FetchLatest returns alerts since the given unique ID
	FetchLatest(ctx context.Context, lastID uint64) ([]domain.Alert, error)
}

// Repository defines storage requirements for incidents and events
type Repository interface {
	SaveAlert(ctx context.Context, alert domain.Alert) error
	GetIncidents(ctx context.Context) ([]domain.Incident, error)
	GetLastProcessedID(ctx context.Context) (uint64, error)
	SetLastProcessedID(ctx context.Context, id uint64) error
}

// TimelineService defines the interface for generating outputs
type TimelineService interface {
	Generate(incident domain.Incident) (string, error)
}
