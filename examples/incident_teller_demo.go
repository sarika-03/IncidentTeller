package main

import (
	"fmt"
	"strings"
	"time"

	"incident-teller/internal/domain"
	"incident-teller/internal/services"
)

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║              INCIDENTTELLER - NARRATIVE MODE                   ║")
	fmt.Println("║         (Calm on-call engineer explaining the incident)        ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝\n")

	// Demo different incident scenarios
	scenarios := []struct {
		name   string
		alerts []domain.Alert
	}{
		{
			name:   "Database Memory Leak Cascade",
			alerts: createDatabaseMemoryLeakScenario(),
		},
		{
			name:   "Disk Space Exhaustion",
			alerts: createDiskSpaceScenario(),
		},
		{
			name:   "Network Saturation",
			alerts: createNetworkSaturationScenario(),
		},
	}

	teller := services.NewIncidentTeller()

	for i, scenario := range scenarios {
		if i > 0 {
			fmt.Println("\n\n" + strings.Repeat("═", 64))
			fmt.Println()
		}

		fmt.Printf("SCENARIO %d: %s\n", i+1, scenario.name)
		fmt.Println(strings.Repeat("─", 64))
		fmt.Println()

		// Generate the incident story
		story := teller.TellStory(scenario.alerts)

		// Print the formatted story
		fmt.Print(services.FormatIncidentStory(story))
	}

	fmt.Println("\n\n╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                  ALL SCENARIOS ANALYZED                        ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
}

// Scenario 1: Database memory leak causes cascade
func createDatabaseMemoryLeakScenario() []domain.Alert {
	baseTime := time.Now().Add(-25 * time.Minute)

	return []domain.Alert{
		{
			ID:          "alert-001",
			Name:        "postgres_memory_usage",
			Status:      domain.StatusWarning,
			OldStatus:   domain.StatusClear,
			ResourceType: domain.ResourceMemory,
			Chart:       "apps.postgres.memory",
			Host:        "db-primary-01",
			Value:       78.2,
			OccurredAt:  baseTime,
			Description: "PostgreSQL memory climbing",
		},
		{
			ID:          "alert-002",
			Name:        "system_memory_critical",
			Status:      domain.StatusCritical,
			OldStatus:   domain.StatusWarning,
			ResourceType: domain.ResourceMemory,
			Chart:       "system.ram",
			Host:        "db-primary-01",
			Value:       94.5,
			OccurredAt:  baseTime.Add(5 * time.Minute),
			Description: "System memory exhausted",
		},
		{
			ID:          "alert-003",
			Name:        "swap_usage_high",
			Status:      domain.StatusCritical,
			OldStatus:   domain.StatusClear,
			ResourceType: domain.ResourceDisk,
			Chart:       "system.swap",
			Host:        "db-primary-01",
			Value:       88.7,
			OccurredAt:  baseTime.Add(8 * time.Minute),
			Description: "Heavy swapping started",
		},
		{
			ID:          "alert-004",
			Name:        "cpu_iowait",
			Status:      domain.StatusCritical,
			OldStatus:   domain.StatusClear,
			ResourceType: domain.ResourceCPU,
			Chart:       "system.cpu",
			Host:        "db-primary-01",
			Value:       67.3,
			OccurredAt:  baseTime.Add(10 * time.Minute),
			Description: "CPU stuck in iowait",
		},
		{
			ID:          "alert-005",
			Name:        "query_latency_high",
			Status:      domain.StatusWarning,
			OldStatus:   domain.StatusClear,
			ResourceType: domain.ResourceNetwork,
			Chart:       "apps.postgres.latency",
			Host:        "db-primary-01",
			Value:       1250.0,
			OccurredAt:  baseTime.Add(12 * time.Minute),
			Description: "Query response time degraded",
		},
	}
}

// Scenario 2: Disk space exhaustion
func createDiskSpaceScenario() []domain.Alert {
	baseTime := time.Now().Add(-15 * time.Minute)

	return []domain.Alert{
		{
			ID:          "disk-001",
			Name:        "disk_space_warning",
			Status:      domain.StatusWarning,
			OldStatus:   domain.StatusClear,
			ResourceType: domain.ResourceDisk,
			Chart:       "system.disk.root",
			Host:        "web-01",
			Value:       82.5,
			OccurredAt:  baseTime,
			Description: "Root partition filling up",
		},
		{
			ID:          "disk-002",
			Name:        "disk_space_critical",
			Status:      domain.StatusCritical,
			OldStatus:   domain.StatusWarning,
			ResourceType: domain.ResourceDisk,
			Chart:       "system.disk.root",
			Host:        "web-01",
			Value:       95.8,
			OccurredAt:  baseTime.Add(3 * time.Minute),
			Description: "Root partition nearly full",
		},
		{
			ID:          "disk-003",
			Name:        "log_write_errors",
			Status:      domain.StatusCritical,
			OldStatus:   domain.StatusClear,
			ResourceType: domain.ResourceProcess,
			Chart:       "apps.logger",
			Host:        "web-01",
			Value:       100.0,
			OccurredAt:  baseTime.Add(4 * time.Minute),
			Description: "Cannot write logs - disk full",
		},
		{
			ID:          "disk-004",
			Name:        "api_failures",
			Status:      domain.StatusCritical,
			OldStatus:   domain.StatusClear,
			ResourceType: domain.ResourceNetwork,
			Chart:       "apps.api.errors",
			Host:        "web-01",
			Value:       45.0,
			OccurredAt:  baseTime.Add(5 * time.Minute),
			Description: "API error rate spiking",
		},
	}
}

// Scenario 3: Network saturation
func createNetworkSaturationScenario() []domain.Alert {
	baseTime := time.Now().Add(-10 * time.Minute)

	return []domain.Alert{
		{
			ID:          "net-001",
			Name:        "high_network_traffic",
			Status:      domain.StatusWarning,
			OldStatus:   domain.StatusClear,
			ResourceType: domain.ResourceNetwork,
			Chart:       "system.net.bandwidth",
			Host:        "app-server-03",
			Value:       750.0,
			OccurredAt:  baseTime,
			Description: "Network bandwidth at 75%",
		},
		{
			ID:          "net-002",
			Name:        "connection_timeouts",
			Status:      domain.StatusCritical,
			OldStatus:   domain.StatusWarning,
			ResourceType: domain.ResourceNetwork,
			Chart:       "apps.api.connections",
			Host:        "app-server-03",
			Value:       125.0,
			OccurredAt:  baseTime.Add(2 * time.Minute),
			Description: "Connection timeout rate increased",
		},
		{
			ID:          "net-003",
			Name:        "cpu_spike",
			Status:      domain.StatusWarning,
			OldStatus:   domain.StatusClear,
			ResourceType: domain.ResourceCPU,
			Chart:       "system.cpu",
			Host:        "app-server-03",
			Value:       72.0,
			OccurredAt:  baseTime.Add(3 * time.Minute),
			Description: "CPU handling network overhead",
		},
	}
}
