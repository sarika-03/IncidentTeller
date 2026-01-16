package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"incident-teller/internal/domain"
	"incident-teller/internal/services"
)

func main() {
	fmt.Println("🧪 IncidentTeller Analysis Demo")
	fmt.Println("================================\n")

	// Create analyzer
	analyzer := services.NewIncidentAnalyzer()

	// Simulate a real incident scenario: Memory leak → Disk swap → CPU iowait
	now := time.Now()
	
	alerts := []domain.Alert{
		// T+0s: Process memory leak starts
		{
			ID:           "alert-001",
			ExternalID:   1001,
			Host:         "prod-web-01",
			Chart:        "apps.mem",
			Family:       "mem",
			Name:         "app_memory_usage",
			Status:       domain.StatusWarning,
			OldStatus:    domain.StatusClear,
			Value:        75.5,
			OccurredAt:   now,
			Description:  "Application memory usage above threshold",
			ResourceType: domain.ResourceMemory,
		},
		// T+30s: Memory escalates to critical
		{
			ID:           "alert-002",
			ExternalID:   1002,
			Host:         "prod-web-01",
			Chart:        "system.ram",
			Family:       "mem",
			Name:         "system_memory_usage",
			Status:       domain.StatusCritical,
			OldStatus:    domain.StatusWarning,
			Value:        92.3,
			OccurredAt:   now.Add(30 * time.Second),
			Description:  "System memory critically low",
			ResourceType: domain.ResourceMemory,
		},
		// T+45s: Disk swap starts happening
		{
			ID:           "alert-003",
			ExternalID:   1003,
			Host:         "prod-web-01",
			Chart:        "system.swap",
			Family:       "disk",
			Name:         "swap_usage",
			Status:       domain.StatusWarning,
			OldStatus:    domain.StatusClear,
			Value:        45.0,
			OccurredAt:   now.Add(45 * time.Second),
			Description:  "Swap usage increasing",
			ResourceType: domain.ResourceDisk,
		},
		// T+60s: CPU iowait spikes
		{
			ID:           "alert-004",
			ExternalID:   1004,
			Host:         "prod-web-01",
			Chart:        "system.cpu",
			Family:       "cpu",
			Name:         "cpu_iowait",
			Status:       domain.StatusCritical,
			OldStatus:    domain.StatusClear,
			Value:        78.5,
			OccurredAt:   now.Add(60 * time.Second),
			Description:  "High CPU iowait detected",
			ResourceType: domain.ResourceCPU,
		},
		// T+90s: Network starts degrading due to buffer pressure
		{
			ID:           "alert-005",
			ExternalID:   1005,
			Host:         "prod-web-01",
			Chart:        "net.drops",
			Family:       "network",
			Name:         "packet_drops",
			Status:       domain.StatusWarning,
			OldStatus:    domain.StatusClear,
			Value:        150.0,
			OccurredAt:   now.Add(90 * time.Second),
			Description:  "Network packet drops detected",
			ResourceType: domain.ResourceNetwork,
		},
		// T+120s: Memory clears (OOM killer or app restart)
		{
			ID:           "alert-006",
			ExternalID:   1006,
			Host:         "prod-web-01",
			Chart:        "apps.mem",
			Family:       "mem",
			Name:         "app_memory_usage",
			Status:       domain.StatusClear,
			OldStatus:    domain.StatusCritical,
			Value:        25.0,
			OccurredAt:   now.Add(120 * time.Second),
			Description:  "Application memory usage normalized",
			ResourceType: domain.ResourceMemory,
		},
		// T+135s: Disk clears
		{
			ID:           "alert-007",
			ExternalID:   1007,
			Host:         "prod-web-01",
			Chart:        "system.swap",
			Family:       "disk",
			Name:         "swap_usage",
			Status:       domain.StatusClear,
			OldStatus:    domain.StatusWarning,
			Value:        5.0,
			OccurredAt:   now.Add(135 * time.Second),
			Description:  "Swap usage back to normal",
			ResourceType: domain.ResourceDisk,
		},
		// T+150s: CPU clears
		{
			ID:           "alert-008",
			ExternalID:   1008,
			Host:         "prod-web-01",
			Chart:        "system.cpu",
			Family:       "cpu",
			Name:         "cpu_iowait",
			Status:       domain.StatusClear,
			OldStatus:    domain.StatusCritical,
			Value:        5.0,
			OccurredAt:   now.Add(150 * time.Second),
			Description:  "CPU iowait normalized",
			ResourceType: domain.ResourceCPU,
		},
	}

	log.Printf("📊 Analyzing %d alert events...\n\n", len(alerts))

	// Analyze the incident
	timeline := analyzer.AnalyzeIncident(alerts)

	// Generate and display summary
	summary := analyzer.GenerateIncidentSummary(timeline)
	fmt.Println(summary)

	// Detailed timeline with causation
	fmt.Println("\n=== Detailed Causality Analysis ===\n")
	for i, entry := range timeline {
		fmt.Printf("%d. [%s] %s\n", i+1, entry.Timestamp.Format("15:04:05"), entry.Type)
		fmt.Printf("   Resource: %s | Severity: %s\n", entry.ResourceType, entry.Severity)
		fmt.Printf("   %s\n", entry.Message)
		
		if len(entry.CausedBy) > 0 {
			fmt.Printf("   ⚡ Triggered by: %v\n", entry.CausedBy)
		}
		
		if entry.DurationSinceStart != nil {
			fmt.Printf("   ⏱️  +%s from incident start\n", entry.DurationSinceStart.Round(time.Second))
		}
		fmt.Println()
	}

	fmt.Println("\n✅ Demo complete!")
}
