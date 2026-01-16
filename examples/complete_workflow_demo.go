package main

import (
	"fmt"
	"time"

	"incident-teller/internal/domain"
	"incident-teller/internal/services"
)

// This example demonstrates IncidentTeller's complete workflow:
// 1. Root cause analysis with confidence scoring
// 2. Blast radius with direct/indirect/unaffected classification
// 3. Actionable fixes (immediate, short-term, long-term)
// 4. Narrative storytelling in calm engineer voice

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║        INCIDENTTELLER - COMPLETE WORKFLOW DEMO                ║")
	fmt.Println("║                                                                ║")
	fmt.Println("║  Scenario: Production database memory leak causes cascade     ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝\n")

	// Simulate a REAL production incident
	alerts := createProductionIncident()

	fmt.Println("📥 Ingested", len(alerts), "alerts from monitoring system\n")
	fmt.Println("🔄 Analyzing incident...\n")

	// Initialize IncidentTeller
	teller := services.NewIncidentTeller()

	// Generate the incident story
	story := teller.TellStory(alerts)

	// Display the narrative report
	fmt.Println(services.FormatIncidentStory(story))

	// Show detailed technical analysis
	fmt.Println("\n╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║          DETAILED TECHNICAL ANALYSIS (for SRE team)           ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝\n")

	analyzer := services.NewComprehensiveIncidentAnalyzer()
	intelligence := analyzer.Analyze(alerts)

	// Show root cause comparison
	fmt.Println("🎯 ROOT CAUSE CANDIDATES:\n")
	fmt.Printf("Primary: %s (Confidence: %d/100)\n", 
		intelligence.RootCause.Alert.Name,
		intelligence.RootCause.ConfidenceScore)
	
	fmt.Println("\nWhy this is most likely:")
	for _, evidence := range intelligence.RootCause.Evidence {
		fmt.Printf("  ✓ %s\n", evidence)
	}

	if len(intelligence.AlternativeCauses) > 0 {
		fmt.Println("\nAlternatives considered but less likely:")
		for i, alt := range intelligence.AlternativeCauses {
			if i >= 3 {
				break
			}
			gap := intelligence.RootCause.ConfidenceScore - alt.ConfidenceScore
			fmt.Printf("  %d. %s (Confidence: %d/100, -%d points)\n",
				i+1, alt.Alert.Name, alt.ConfidenceScore, gap)
		}
	}

	// Show blast radius details
	fmt.Println("\n💥 BLAST RADIUS BREAKDOWN:\n")
	severityLabel := "UNKNOWN"
	switch {
	case intelligence.BlastRadius.ImpactScore >= 80:
		severityLabel = "CRITICAL"
	case intelligence.BlastRadius.ImpactScore >= 60:
		severityLabel = "HIGH"
	case intelligence.BlastRadius.ImpactScore >= 40:
		severityLabel = "MEDIUM"
	default:
		severityLabel = "LOW"
	}
	fmt.Printf("Impact Score: %d/100 (%s severity)\n",
		intelligence.BlastRadius.ImpactScore,
		severityLabel)
	fmt.Printf("Cascade Depth: %d levels\n", intelligence.BlastRadius.CascadeDepth)
	fmt.Printf("Recovery Estimate: %s\n\n", intelligence.BlastRadius.RecoveryEstimate)

	fmt.Printf("Directly Affected: %d components\n", 
		len(intelligence.BlastRadius.DirectlyAffected))
	fmt.Printf("Indirectly Affected: %d components (cascade)\n",
		len(intelligence.BlastRadius.IndirectlyAffected))
	fmt.Printf("Unaffected: %d components\n",
		len(intelligence.BlastRadius.Unaffected))

	// Show what to do RIGHT NOW
	fmt.Println("\n🚨 IMMEDIATE ACTIONS (copy-paste ready):\n")
	for i, action := range intelligence.ActionableFixes.ImmediateFix {
		if i >= 5 {
			break
		}
		// Skip metadata lines (those starting with emojis)
		if len(action) > 0 && !strings.HasPrefix(action, "⚠") && 
		   !strings.HasPrefix(action, "🎯") && !strings.HasPrefix(action, "🚨") {
			fmt.Printf("  %d. %s\n", i+1, action)
		}
	}

	// Simulation of what a Slack message would look like
	fmt.Println("\n📱 SLACK NOTIFICATION (would be posted to #incidents):\n")
	fmt.Println("────────────────────────────────────────────────────────────────")
	fmt.Print(analyzer.GenerateSlackMessage(intelligence))
	fmt.Println("────────────────────────────────────────────────────────────────")

	fmt.Println("\n✅ Analysis complete! All outputs generated successfully.")
	fmt.Println("\n💡 In production, this would:")
	fmt.Println("   • Post to Slack #incidents channel")
	fmt.Println("   • Create PagerDuty incident with analysis")
	fmt.Println("   • Update status page")
	fmt.Println("   • Log to incident management system")
}

func createProductionIncident() []domain.Alert {
	// Real-world scenario: Database memory leak causes full system cascade
	baseTime := time.Now().Add(-30 * time.Minute)

	return []domain.Alert{
		// Phase 1: Initial memory warning (T+0)
		{
			ID:          "alert-db-mem-001",
			Name:        "postgres_memory_high",
			Status:      domain.StatusWarning,
			OldStatus:   domain.StatusClear,
			ResourceType: domain.ResourceMemory,
			Chart:       "apps.postgres.memory",
			Host:        "db-primary-01",
			Value:       76.3,
			OccurredAt:  baseTime,
			Description: "PostgreSQL shared_buffers + work_mem climbing",
		},

		// Phase 2: Memory reaches critical (T+5min)
		{
			ID:          "alert-db-mem-002",
			Name:        "postgres_memory_critical",
			Status:      domain.StatusCritical,
			OldStatus:   domain.StatusWarning,
			ResourceType: domain.ResourceMemory,
			Chart:       "apps.postgres.memory",
			Host:        "db-primary-01",
			Value:       94.8,
			OccurredAt:  baseTime.Add(5 * time.Minute),
			Description: "Memory leak detected - query cache not releasing",
		},

		// Phase 3: Cascade begins - system memory exhausted (T+7min)
		{
			ID:          "alert-sys-mem-001",
			Name:        "system_memory_critical",
			Status:      domain.StatusCritical,
			OldStatus:   domain.StatusWarning,
			ResourceType: domain.ResourceMemory,
			Chart:       "system.ram",
			Host:        "db-primary-01",
			Value:       97.2,
			OccurredAt:  baseTime.Add(7 * time.Minute),
			Description: "System RAM exhausted - OOM killer about to trigger",
		},

		// Phase 4: Swap thrashing starts (T+9min)
		{
			ID:          "alert-swap-001",
			Name:        "swap_usage_critical",
			Status:      domain.StatusCritical,
			OldStatus:   domain.StatusClear,
			ResourceType: domain.ResourceDisk,
			Chart:       "system.swap",
			Host:        "db-primary-01",
			Value:       91.5,
			OccurredAt:  baseTime.Add(9 * time.Minute),
			Description: "Heavy swap usage - disk I/O saturated",
		},

		// Phase 5: CPU iowait spike (T+11min)
		{
			ID:          "alert-cpu-001",
			Name:        "cpu_iowait_high",
			Status:      domain.StatusCritical,
			OldStatus:   domain.StatusClear,
			ResourceType: domain.ResourceCPU,
			Chart:       "system.cpu.iowait",
			Host:        "db-primary-01",
			Value:       68.4,
			OccurredAt:  baseTime.Add(11 * time.Minute),
			Description: "CPU spending 68% waiting on swap I/O",
		},

		// Phase 6: Database queries slow (T+13min)
		{
			ID:          "alert-db-slow-001",
			Name:        "query_latency_p95",
			Status:      domain.StatusCritical,
			OldStatus:   domain.StatusWarning,
			ResourceType: domain.ResourceNetwork,
			Chart:       "apps.postgres.query_latency",
			Host:        "db-primary-01",
			Value:       2850.0,
			OccurredAt:  baseTime.Add(13 * time.Minute),
			Description: "P95 query latency: 2.85s (normal: 50ms)",
		},

		// Phase 7: Application tier degradation (T+15min)
		{
			ID:          "alert-app-001",
			Name:        "api_response_time",
			Status:      domain.StatusCritical,
			OldStatus:   domain.StatusWarning,
			ResourceType: domain.ResourceNetwork,
			Chart:       "apps.api.response_time",
			Host:        "app-server-01",
			Value:       4200.0,
			OccurredAt:  baseTime.Add(15 * time.Minute),
			Description: "API endpoints timing out waiting for DB",
		},

		// Phase 8: Connection pool exhausted (T+16min)
		{
			ID:          "alert-db-conn-001",
			Name:        "connection_pool_exhausted",
			Status:      domain.StatusCritical,
			OldStatus:   domain.StatusWarning,
			ResourceType: domain.ResourceProcess,
			Chart:       "apps.postgres.connections",
			Host:        "db-primary-01",
			Value:       100.0,
			OccurredAt:  baseTime.Add(16 * time.Minute),
			Description: "All 100 connections in use, new requests queuing",
		},

		// Phase 9: User-facing errors (T+18min)
		{
			ID:          "alert-app-errors-001",
			Name:        "http_5xx_errors",
			Status:      domain.StatusCritical,
			OldStatus:   domain.StatusClear,
			ResourceType: domain.ResourceNetwork,
			Chart:       "apps.api.errors",
			Host:        "app-server-01",
			Value:       35.0,
			OccurredAt:  baseTime.Add(18 * time.Minute),
			Description: "35% of requests failing with 503/504 errors",
		},
	}
}
