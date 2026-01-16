package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"incident-teller/internal/domain"
	"incident-teller/internal/ports"
)

// RealTimePoller continuously polls Netdata for new alerts
type RealTimePoller struct {
	source       ports.AlertSource
	repository   ports.Repository
	analyzer     *IncidentAnalyzer
	pollInterval time.Duration
	eventChan    chan []domain.Alert
}

// NewRealTimePoller creates a new real-time alert poller
func NewRealTimePoller(
	source ports.AlertSource,
	repo ports.Repository,
	analyzer *IncidentAnalyzer,
	pollInterval time.Duration,
) *RealTimePoller {
	return &RealTimePoller{
		source:       source,
		repository:   repo,
		analyzer:     analyzer,
		pollInterval: pollInterval,
		eventChan:    make(chan []domain.Alert, 100),
	}
}

// Start begins the polling loop
func (p *RealTimePoller) Start(ctx context.Context) error {
	log.Println("🚀 Starting real-time alert poller...")

	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("⏹️  Poller stopped")
			return ctx.Err()
		case <-ticker.C:
			if err := p.poll(ctx); err != nil {
				log.Printf("⚠️  Poll error: %v", err)
				// Continue polling even on error
			}
		}
	}
}

// poll fetches and processes new alerts
func (p *RealTimePoller) poll(ctx context.Context) error {
	// Get last processed ID
	lastID, err := p.repository.GetLastProcessedID(ctx)
	if err != nil {
		log.Printf("Failed to get last processed ID (using 0): %v", err)
		lastID = 0
	}

	// Fetch new alerts
	alerts, err := p.source.FetchLatest(ctx, lastID)
	if err != nil {
		return fmt.Errorf("failed to fetch alerts: %w", err)
	}

	if len(alerts) == 0 {
		return nil // No new alerts
	}

	log.Printf("📥 Received %d new alerts", len(alerts))

	// Save alerts
	var maxID uint64
	for _, alert := range alerts {
		if err := p.repository.SaveAlert(ctx, alert); err != nil {
			log.Printf("⚠️  Failed to save alert %s: %v", alert.ID, err)
			continue
		}

		if alert.ExternalID > maxID {
			maxID = alert.ExternalID
		}
	}

	// Update last processed ID
	if maxID > 0 {
		if err := p.repository.SetLastProcessedID(ctx, maxID); err != nil {
			log.Printf("⚠️  Failed to update last processed ID: %v", err)
		}
	}

	// Send to event channel for consumers
	select {
	case p.eventChan <- alerts:
	default:
		log.Println("⚠️  Event channel full, dropping alerts")
	}

	// Analyze and log
	timeline := p.analyzer.AnalyzeIncident(alerts)
	if len(timeline) > 0 {
		log.Println("📊 Incident Analysis:")
		for _, entry := range timeline {
			log.Printf("  [%s] %s - %s",
				entry.Timestamp.Format("15:04:05"),
				entry.Type,
				entry.Message)
		}
	}

	return nil
}

// Events returns the channel for consuming alert events
func (p *RealTimePoller) Events() <-chan []domain.Alert {
	return p.eventChan
}

// PollOnce performs a single poll (useful for testing or manual triggers)
func (p *RealTimePoller) PollOnce(ctx context.Context) ([]domain.Alert, error) {
	lastID, err := p.repository.GetLastProcessedID(ctx)
	if err != nil {
		lastID = 0
	}

	alerts, err := p.source.FetchLatest(ctx, lastID)
	if err != nil {
		return nil, err
	}

	// Save and update
	var maxID uint64
	for _, alert := range alerts {
		if err := p.repository.SaveAlert(ctx, alert); err != nil {
			continue
		}
		if alert.ExternalID > maxID {
			maxID = alert.ExternalID
		}
	}

	if maxID > 0 {
		p.repository.SetLastProcessedID(ctx, maxID)
	}

	return alerts, nil
}
