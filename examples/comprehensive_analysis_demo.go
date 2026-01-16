package main

import (
	"fmt"
	"time"

	"incident-teller/internal/domain"
	"incident-teller/internal/services"
)

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║   COMPREHENSIVE INCIDENT ANALYSIS - ALL OUTPUT FORMATS         ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝\n")

	// Create incident scenario: Database memory leak causes cascade
	alerts := createIncidentScenario()

	// Initialize comprehensive analyzer
	analyzer := services.NewComprehensiveIncidentAnalyzer()

	// Perform complete analysis
	fmt.Println("🔄 Analyzing incident...")
	intelligence := analyzer.Analyze(alerts)
	fmt.Println("✅ Analysis complete!\n")

	// Output 1: Executive Summary (for leadership / PagerDuty)
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println("OUTPUT 1: EXECUTIVE SUMMARY (for Leadership)")
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Print(analyzer.GenerateExecutiveSummary(intelligence))
	fmt.Println()

	// Output 2: Technical Report (for on-call engineers)
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println("OUTPUT 2: TECHNICAL REPORT (for On-Call Engineers)")
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Print(analyzer.GenerateTechnicalReport(intelligence))
	fmt.Println()

	// Output 3: Slack Message (for incident channel)
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println("OUTPUT 3: SLACK MESSAGE (for #incidents channel)")
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println(analyzer.GenerateSlackMessage(intelligence))
	fmt.Println()

	// Output 4: Detailed Component Impact (for infrastructure team)
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println("OUTPUT 4: COMPONENT IMPACT ANALYSIS (for Infrastructure Team)")
	fmt.Println("════════════════════════════════════════════════════════════════")
	printComponentImpactAnalysis(intelligence)
	fmt.Println()

	// Output 5: Comparative Root Cause Analysis
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println("OUTPUT 5: WHY THIS ROOT CAUSE VS ALTERNATIVES")
	fmt.Println("════════════════════════════════════════════════════════════════")
	printComparativeAnalysis(intelligence)
	fmt.Println()

	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    ANALYSIS COMPLETE                           ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝")
}

func createIncidentScenario() []domain.Alert {
	return []domain.Alert{
		{
			ID:          "alert-001",
			Name:        "postgres_memory_high",
			Status:      domain.StatusWarning,
			OldStatus:   domain.StatusClear,
			ResourceType: domain.ResourceMemory,
			Chart:       "apps.postgres.memory",
			Host:        "db-primary-01",
			Value:       78.5,
			OccurredAt:  time.Now().Add(-20 * time.Minute),
			Description: "PostgreSQL memory consumption increasing",
		},
		{
			ID:          "alert-002",
			Name:        "postgres_memory_critical",
			Status:      domain.StatusCritical,
			OldStatus:   domain.StatusWarning,
			ResourceType: domain.ResourceMemory,
			Chart:       "apps.postgres.memory",
			Host:        "db-primary-01",
			Value:       96.8,
			OccurredAt:  time.Now().Add(-15 * time.Minute),
			Description: "PostgreSQL consuming excessive memory - possible leak",
		},
		{
			ID:          "alert-003",
			Name:        "system_swap_usage",
			Status:      domain.StatusCritical,
			OldStatus:   domain.StatusClear,
			ResourceType: domain.ResourceDisk,
			Chart:       "system.swap",
			Host:        "db-primary-01",
			Value:       92.3,
			OccurredAt:  time.Now().Add(-12 * time.Minute),
			Description: "System swapping heavily - performance degraded",
		},
		{
			ID:          "alert-004",
			Name:        "cpu_iowait_high",
			Status:      domain.StatusCritical,
			OldStatus:   domain.StatusClear,
			ResourceType: domain.ResourceCPU,
			Chart:       "system.cpu.iowait",
			Host:        "db-primary-01",
			Value:       72.4,
			OccurredAt:  time.Now().Add(-10 * time.Minute),
			Description: "High CPU iowait due to swap thrashing",
		},
		{
			ID:          "alert-005",
			Name:        "db_query_latency",
			Status:      domain.StatusWarning,
			OldStatus:   domain.StatusClear,
			ResourceType: domain.ResourceNetwork,
			Chart:       "apps.postgres.query_time",
			Host:        "db-primary-01",
			Value:       850.0,
			OccurredAt:  time.Now().Add(-8 * time.Minute),
			Description: "Database query latency increased to 850ms",
		},
		{
			ID:          "alert-006",
			Name:        "api_response_time",
			Status:      domain.StatusWarning,
			OldStatus:   domain.StatusClear,
			ResourceType: domain.ResourceNetwork,
			Chart:       "apps.api.response_time",
			Host:        "api-server-01",
			Value:       3200.0,
			OccurredAt:  time.Now().Add(-5 * time.Minute),
			Description: "API response time degraded - database bottleneck",
		},
	}
}

func printComponentImpactAnalysis(intelligence services.IncidentIntelligence) {
	br := intelligence.BlastRadius

	fmt.Printf("📊 Impact Score: %d/100\n", br.ImpactScore)
	fmt.Printf("📝 Summary: %s\n\n", br.SimpleSummary)

	// Directly affected
	fmt.Println("🔴 DIRECTLY AFFECTED COMPONENTS:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	if len(br.DirectlyAffected) == 0 {
		fmt.Println("   None")
	} else {
		for i, comp := range br.DirectlyAffected {
			fmt.Printf("%d. %s (%s)\n", i+1, comp.Name, comp.Type)
			if comp.AffectedAt != nil {
				fmt.Printf("   🕐 First seen: %s\n", comp.AffectedAt.Format("15:04:05"))
			}
			if len(comp.MetricValues) > 0 {
				fmt.Printf("   📊 Peak value: %.2f\n", maxValue(comp.MetricValues))
			}
			if len(comp.Evidence) > 0 {
				fmt.Printf("   📋 Why classified as direct:\n")
				for _, ev := range comp.Evidence {
					fmt.Printf("      • %s\n", ev)
				}
			}
			fmt.Println()
		}
	}

	// Indirectly affected
	fmt.Println("🟡 INDIRECTLY AFFECTED (Cascade Effects):")
	fmt.Println("───────────────────────────────────────────────────────────────")
	if len(br.IndirectlyAffected) == 0 {
		fmt.Println("   None - no cascade detected")
	} else {
		fmt.Printf("   Cascade Depth: %d levels\n\n", br.CascadeDepth)
		for i, comp := range br.IndirectlyAffected {
			fmt.Printf("%d. %s (%s)\n", i+1, comp.Name, comp.Type)
			if comp.AffectedAt != nil {
				fmt.Printf("   🕐 Affected at: %s\n", comp.AffectedAt.Format("15:04:05"))
			}
			if len(comp.Evidence) > 0 {
				fmt.Printf("   📋 Why classified as indirect:\n")
				for _, ev := range comp.Evidence {
					fmt.Printf("      • %s\n", ev)
				}
			}
			fmt.Println()
		}
	}

	// Unaffected
	fmt.Println("🟢 UNAFFECTED COMPONENTS:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	if len(br.Unaffected) == 0 {
		fmt.Println("   All monitored systems are affected!")
	} else {
		for i, comp := range br.Unaffected {
			fmt.Printf("%d. %s - %s\n", i+1, comp.Name, comp.Evidence[0])
		}
	}
}

func printComparativeAnalysis(intelligence services.IncidentIntelligence) {
	fmt.Printf("🎯 Primary Root Cause: %s\n", intelligence.RootCause.Alert.Name)
	fmt.Printf("   Confidence Score: %d/100\n", intelligence.RootCause.ConfidenceScore)
	fmt.Printf("   Overall Assessment: %s\n\n", intelligence.ConfidenceLevel)

	if len(intelligence.AlternativeCauses) == 0 {
		fmt.Println("No alternative causes identified - single clear root cause.")
		return
	}

	fmt.Println("📊 Comparison with Alternative Causes:")
	fmt.Println("───────────────────────────────────────────────────────────────\n")

	topScore := intelligence.RootCause.ConfidenceScore

	for i, alt := range intelligence.AlternativeCauses {
		gap := topScore - alt.ConfidenceScore
		gapPercent := float64(gap) / float64(topScore) * 100

		fmt.Printf("%d. %s\n", i+1, alt.Alert.Name)
		fmt.Printf("   Score: %d/100 (-%d from primary, %.1f%% lower)\n", 
			alt.ConfidenceScore, gap, gapPercent)
		
		// Explain why it scored lower
		if gap >= 30 {
			fmt.Println("   ❌ Significantly less likely:")
		} else if gap >= 15 {
			fmt.Println("   ⚠️ Moderately less likely:")
		} else {
			fmt.Println("   ⚡ Very close - could be contributing factor:")
		}

		// Compare evidence
		primaryEvidence := len(intelligence.RootCause.Evidence)
		altEvidence := len(alt.Evidence)
		fmt.Printf("      • Evidence count: %d vs %d (primary)\n", altEvidence, primaryEvidence)
		
		if !alt.IsEarliest && intelligence.RootCause.IsEarliest {
			fmt.Println("      • Occurred later in timeline (not first failure)")
		}
		
		if !alt.HasCascade && intelligence.RootCause.HasCascade {
			fmt.Println("      • No cascading effects detected")
		}
		
		if !alt.HasLogErrors && intelligence.RootCause.HasLogErrors {
			fmt.Println("      • No correlated error logs")
		}

		fmt.Println()
	}

	// Final recommendation
	fmt.Println("💡 RECOMMENDATION:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	
	if topScore >= 80 {
		closeCalls := 0
		for _, alt := range intelligence.AlternativeCauses {
			if topScore-alt.ConfidenceScore < 15 {
				closeCalls++
			}
		}
		
		if closeCalls > 0 {
			fmt.Printf("Focus on primary cause (%s) but monitor %d close alternatives\n",
				intelligence.RootCause.Alert.Name, closeCalls)
		} else {
			fmt.Printf("High confidence - focus exclusively on %s\n",
				intelligence.RootCause.Alert.Name)
		}
	} else if topScore >= 60 {
		fmt.Println("Moderate confidence - address primary cause while investigating top alternative")
	} else {
		fmt.Println("Multiple viable causes - investigate top 2-3 in parallel")
	}
}

func maxValue(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values[1:] {
		if v > max {
			max = v
		}
	}
	return max
}
