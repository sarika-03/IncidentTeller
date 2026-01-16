package main

import (
	"fmt"
	"time"

	"incident-teller/internal/domain"
	"incident-teller/internal/services"
)

// simulateIncident creates a realistic incident scenario for demo
func simulateIncident() []domain.Alert {
	baseTime := time.Now().Add(-30 * time.Minute)

	return []domain.Alert{
		// 1. Root cause: Memory leak in application process
		{
			ID:           "host1-1001",
			ExternalID:   1001,
			Host:         "web-server-01",
			Chart:        "apps.mem",
			Family:       "mem",
			Name:         "app_memory_usage",
			Status:       domain.StatusWarning,
			OldStatus:    domain.StatusClear,
			Value:        75.5,
			OccurredAt:   baseTime,
			Description:  "Application memory usage above 70%",
			ResourceType: domain.ResourceMemory,
		},
		// 2. Memory escalates to critical
		{
			ID:           "host1-1002",
			ExternalID:   1002,
			Host:         "web-server-01",
			Chart:        "apps.mem",
			Family:       "mem",
			Name:         "app_memory_usage",
			Status:       domain.StatusCritical,
			OldStatus:    domain.StatusWarning,
			Value:        92.3,
			OccurredAt:   baseTime.Add(3 * time.Minute),
			Description:  "Application memory usage above 90%",
			ResourceType: domain.ResourceMemory,
		},
		// 3. Cascade: Disk starts thrashing (swap usage)
		{
			ID:           "host1-1003",
			ExternalID:   1003,
			Host:         "web-server-01",
			Chart:        "system.swap",
			Family:       "disk",
			Name:         "swap_usage",
			Status:       domain.StatusWarning,
			OldStatus:    domain.StatusClear,
			Value:        65.0,
			OccurredAt:   baseTime.Add(4 * time.Minute),
			Description:  "High swap usage detected",
			ResourceType: domain.ResourceDisk,
		},
		// 4. Cascade: CPU iowait increases
		{
			ID:           "host1-1004",
			ExternalID:   1004,
			Host:         "web-server-01",
			Chart:        "system.cpu",
			Family:       "cpu",
			Name:         "cpu_iowait",
			Status:       domain.StatusWarning,
			OldStatus:    domain.StatusClear,
			Value:        45.2,
			OccurredAt:   baseTime.Add(6 * time.Minute),
			Description:  "High CPU iowait time",
			ResourceType: domain.ResourceCPU,
		},
		// 5. CPU escalates to critical
		{
			ID:           "host1-1005",
			ExternalID:   1005,
			Host:         "web-server-01",
			Chart:        "system.cpu",
			Family:       "cpu",
			Name:         "cpu_usage",
			Status:       domain.StatusCritical,
			OldStatus:    domain.StatusWarning,
			Value:        95.7,
			OccurredAt:   baseTime.Add(8 * time.Minute),
			Description:  "Critical CPU usage",
			ResourceType: domain.ResourceCPU,
		},
		// 6. Network degradation (secondary effect)
		{
			ID:           "host1-1006",
			ExternalID:   1006,
			Host:         "web-server-01",
			Chart:        "net.drops",
			Family:       "network",
			Name:         "packet_drops",
			Status:       domain.StatusWarning,
			OldStatus:    domain.StatusClear,
			Value:        1250.0,
			OccurredAt:   baseTime.Add(10 * time.Minute),
			Description:  "Packet drops detected",
			ResourceType: domain.ResourceNetwork,
		},
		// 7. Partial recovery: Memory improves after restart
		{
			ID:           "host1-1007",
			ExternalID:   1007,
			Host:         "web-server-01",
			Chart:        "apps.mem",
			Family:       "mem",
			Name:         "app_memory_usage",
			Status:       domain.StatusClear,
			OldStatus:    domain.StatusCritical,
			Value:        35.2,
			OccurredAt:   baseTime.Add(15 * time.Minute),
			Description:  "Memory usage normalized",
			ResourceType: domain.ResourceMemory,
		},
		// 8. CPU resolves
		{
			ID:           "host1-1008",
			ExternalID:   1008,
			Host:         "web-server-01",
			Chart:        "system.cpu",
			Family:       "cpu",
			Name:         "cpu_usage",
			Status:       domain.StatusClear,
			OldStatus:    domain.StatusCritical,
			Value:        25.3,
			OccurredAt:   baseTime.Add(16 * time.Minute),
			Description:  "CPU usage normalized",
			ResourceType: domain.ResourceCPU,
		},
	}
}

func main() {
	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║    IncidentTeller - SRE Incident Analysis Demo               ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Simulate incident data
	alerts := simulateIncident()
	fmt.Printf("📊 Loaded %d alert events from simulated incident\n\n", len(alerts))

	// Create SRE analyzer
	sreAnalyzer := services.NewSREAnalyzer()

	// Perform analysis
	fmt.Println("🔍 Analyzing incident...")
	fmt.Println()

	explanation := sreAnalyzer.AnalyzeIncidentForSRE(alerts)

	// Display formatted report
	report := services.FormatIncidentExplanation(explanation)
	fmt.Println(report)

	// Additional insights
	fmt.Println("\n🧠 CONFIDENCE SCORING BREAKDOWN")
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("Scoring Factors:")
	fmt.Println("  • Temporal ordering (earlier = higher weight)    : Max 40 pts")
	fmt.Println("  • Cascading resource exhaustion                 : Max 30 pts")
	fmt.Println("  • Alert severity (CRITICAL > WARNING)            : Max 15 pts")
	fmt.Println("  • Related log errors                             : Max 15 pts")
	fmt.Println("  • Resource criticality (Memory/Process)          : Max 10 pts")
	fmt.Println()

	fmt.Println("Why this root cause is most likely:")
	if len(explanation.RootCause.Evidence) > 0 {
		for _, evidence := range explanation.RootCause.Evidence {
			fmt.Printf("  ✓ %s\n", evidence)
		}
	}

	fmt.Println()
	fmt.Println("═══════════════════════════════════════════════════════════════")
	fmt.Println("Analysis complete. Incident report ready for on-call review.")
	fmt.Println()
}
