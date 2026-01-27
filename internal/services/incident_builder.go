package services

import (
	"fmt"
	"sort"
	"time"

	"incident-teller/internal/domain"
)

type IncidentBuilder struct {
	window time.Duration
}

func NewIncidentBuilder(window time.Duration) *IncidentBuilder {
	return &IncidentBuilder{window: window}
}


func (b *IncidentBuilder) Build(alerts []domain.Alert) []domain.Incident {
	if len(alerts) == 0 {
		return nil
	}

	sort.Slice(alerts, func(i, j int) bool {
		return alerts[i].OccurredAt.Before(alerts[j].OccurredAt)
	})

	var incidents []domain.Incident
	if len(alerts) == 0 {
		return nil
	}
	
	current := domain.Incident{
		ID:        fmt.Sprintf("incident-%s-%d", alerts[0].Host, alerts[0].OccurredAt.Unix()),
		StartedAt: alerts[0].OccurredAt,
		Status:    alerts[0].Status,
	}

	for _, alert := range alerts {
		if alert.OccurredAt.Sub(current.StartedAt) > b.window {
			incidents = append(incidents, current)
			current = domain.Incident{
				ID:        fmt.Sprintf("incident-%s-%d", alert.Host, alert.OccurredAt.Unix()),
				StartedAt: alert.OccurredAt,
				Status:    alert.Status,
			}
		}
		current.Events = append(current.Events, alert)
		current.Status = alert.Status
	}

	incidents = append(incidents, current)
	return incidents
}
