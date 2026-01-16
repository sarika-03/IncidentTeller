package services

import (
	"testing"
	"time"

	"incident-teller/internal/domain"
)

func TestIncidentAnalyzer_BasicTimeline(t *testing.T) {
	analyzer := NewIncidentAnalyzer()

	now := time.Now()
	alerts := []domain.Alert{
		{
			ID:           "alert-1",
			Name:         "memory_high",
			Status:       domain.StatusWarning,
			OldStatus:    domain.StatusClear,
			ResourceType: domain.ResourceMemory,
			Host:         "server-01",
			Chart:        "system.ram",
			Value:        85.0,
			OccurredAt:   now,
		},
		{
			ID:           "alert-2",
			Name:         "disk_high",
			Status:       domain.StatusCritical,
			OldStatus:    domain.StatusWarning,
			ResourceType: domain.ResourceDisk,
			Host:         "server-01",
			Chart:        "system.disk",
			Value:        95.0,
			OccurredAt:   now.Add(2 * time.Minute),
		},
	}

	timeline := analyzer.AnalyzeIncident(alerts)

	if len(timeline) != 2 {
		t.Errorf("Expected 2 timeline entries, got %d", len(timeline))
	}

	if timeline[0].Type != "TRIGGERED" {
		t.Errorf("Expected first entry to be TRIGGERED, got %s", timeline[0].Type)
	}
}

func TestIncidentAnalyzer_CascadeDetection(t *testing.T) {
	analyzer := NewIncidentAnalyzer()

	now := time.Now()
	alerts := []domain.Alert{
		{
			ID:           "alert-1",
			Name:         "memory_critical",
			Status:       domain.StatusCritical,
			OldStatus:    domain.StatusWarning,
			ResourceType: domain.ResourceMemory,
			Host:         "server-01",
			Value:        95.0,
			OccurredAt:   now,
		},
		{
			ID:           "alert-2",
			Name:         "swap_high",
			Status:       domain.StatusCritical,
			OldStatus:    domain.StatusClear,
			ResourceType: domain.ResourceDisk,
			Host:         "server-01",
			Value:        90.0,
			OccurredAt:   now.Add(1 * time.Minute),
		},
		{
			ID:           "alert-3",
			Name:         "cpu_iowait",
			Status:       domain.StatusCritical,
			OldStatus:    domain.StatusClear,
			ResourceType: domain.ResourceCPU,
			Host:         "server-01",
			Value:        75.0,
			OccurredAt:   now.Add(3 * time.Minute),
		},
	}

	timeline := analyzer.AnalyzeIncident(alerts)

	// Check if cascade was detected
	foundCascade := false
	for _, entry := range timeline {
		if len(entry.CausedBy) > 0 {
			foundCascade = true
			break
		}
	}

	if !foundCascade {
		t.Error("Expected cascade detection but none found")
	}
}
