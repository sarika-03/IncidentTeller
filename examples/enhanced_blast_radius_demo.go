package main

import (
	"fmt"
	"time"

	"incident-teller/internal/domain"
	"incident-teller/internal/services"
)

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════════════╗")
	fmt.Println("║     ENHANCED SRE INCIDENT ANALYSIS WITH BLAST RADIUS          ║")
	fmt.Println("╚════════════════════════════════════════════════════════════════╝\n")

	// Simulate a realistic production incident: Memory leak causes cascade
	alerts := []domain.Alert{
		{
			ID:          "alert-001",
			Name:        "high_memory_usage",
			Status:      domain.StatusWarning,
			OldStatus:   domain.StatusClear,
			ResourceType: domain.ResourceMemory,
			Chart:       "system.ram",
			Host:        "prod-web-01",
			Value:       82.5,
			OccurredAt:  time.Now().Add(-15 * time.Minute),
			Description: "Application memory usage climbing",
		},
		{
			ID:          "alert-002",
			Name:        "memory_critical",
			Status:      domain.StatusCritical,
			OldStatus:   domain.StatusWarning,
			ResourceType: domain.ResourceMemory,
			Chart:       "system.ram",
			Host:        "prod-web-01",
			Value:       94.2,
			OccurredAt:  time.Now().Add(-12 * time.Minute),
			Description: "Memory exhaustion imminent",
		},
		{
			ID:          "alert-003",
			Name:        "swap_usage_high",
			Status:      domain.StatusCritical,
			OldStatus:   domain.StatusClear,
			ResourceType: domain.ResourceDisk,
			Chart:       "system.swap",
			Host:        "prod-web-01",
			Value:       87.3,
			OccurredAt:  time.Now().Add(-10 * time.Minute),
			Description: "Heavy swapping detected - disk I/O spike",
		},
		{
			ID:          "alert-004",
			Name:        "cpu_iowait",
			Status:      domain.StatusCritical,
			OldStatus:   domain.StatusClear,
			ResourceType: domain.ResourceCPU,
			Chart:       "system.cpu",
			Host:        "prod-web-01",
			Value:       65.8,
			OccurredAt:  time.Now().Add(-8 * time.Minute),
			Description: "CPU waiting on disk I/O",
		},
		{
			ID:          "alert-005",
			Name:        "api_latency",
			Status:      domain.StatusWarning,
			OldStatus:   domain.StatusClear,
			ResourceType: domain.ResourceNetwork,
			Chart:       "app.response_time",
			Host:        "prod-web-01",
			Value:       2500.0,
			OccurredAt:  time.Now().Add(-5 * time.Minute),
			Description: "API response time degraded to 2.5s",
		},
	}

	// 1. Perform root cause analysis
	sreAnalyzer := services.NewSREAnalyzer()
	explanation := sreAnalyzer.AnalyzeIncidentForSRE(alerts)

	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println("STEP 1: ROOT CAUSE IDENTIFICATION")
	fmt.Println("════════════════════════════════════════════════════════════════\n")
	
	fmt.Printf("🎯 Root Cause: %s\n", explanation.RootCause.Alert.Name)
	fmt.Printf("📊 Confidence: %d/100 (%s)\n", 
		explanation.RootCause.ConfidenceScore, 
		explanation.ConfidenceLevel)
	fmt.Printf("📍 Location: %s on %s\n", 
		explanation.RootCause.Alert.Chart, 
		explanation.RootCause.Alert.Host)
	fmt.Printf("💡 Reasoning: %s\n\n", explanation.RootCause.Reasoning)

	// 2. Enhanced blast radius analysis
	blastRadiusAnalyzer := services.NewBlastRadiusAnalyzer()
	enhancedBlastRadius := blastRadiusAnalyzer.AnalyzeBlastRadius(
		alerts, 
		explanation.RootCause,
	)

	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println("STEP 2: BLAST RADIUS ANALYSIS")
	fmt.Println("════════════════════════════════════════════════════════════════\n")

	fmt.Printf("📝 Simple Summary:\n   %s\n\n", enhancedBlastRadius.SimpleSummary)
	fmt.Printf("💥 Impact Score: %d/100\n", enhancedBlastRadius.ImpactScore)
	fmt.Printf("⏱️  Recovery Estimate: %s\n\n", enhancedBlastRadius.RecoveryEstimate)

	// Direct impact
	fmt.Println("🔴 DIRECTLY AFFECTED COMPONENTS:")
	fmt.Println("───────────────────────────────────────────────────────────────")
	if len(enhancedBlastRadius.DirectlyAffected) == 0 {
		fmt.Println("   None")
	} else {
		for i, comp := range enhancedBlastRadius.DirectlyAffected {
			fmt.Printf("%d. %s (%s)\n", i+1, comp.Name, comp.Type)
			if comp.AffectedAt != nil {
				fmt.Printf("   ⏰ Affected at: %s\n", comp.AffectedAt.Format("15:04:05"))
			}
			if len(comp.Evidence) > 0 {
				fmt.Printf("   📋 Evidence:\n")
				for _, ev := range comp.Evidence {
					fmt.Printf("      • %s\n", ev)
				}
			}
			fmt.Println()
		}
	}

	// Indirect impact
	fmt.Println("🟡 INDIRECTLY AFFECTED (CASCADE EFFECTS):")
	fmt.Println("───────────────────────────────────────────────────────────────")
	if len(enhancedBlastRadius.IndirectlyAffected) == 0 {
		fmt.Println("   None - no cascade detected")
	} else {
		for i, comp := range enhancedBlastRadius.IndirectlyAffected {
			fmt.Printf("%d. %s (%s)\n", i+1, comp.Name, comp.Type)
			if comp.AffectedAt != nil {
				fmt.Printf("   ⏰ Affected at: %s\n", comp.AffectedAt.Format("15:04:05"))
			}
			if len(comp.Evidence) > 0 {
				fmt.Printf("   📋 Evidence:\n")
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
	if len(enhancedBlastRadius.Unaffected) == 0 {
		fmt.Println("   All systems affected!")
	} else {
		for i, comp := range enhancedBlastRadius.Unaffected {
			fmt.Printf("%d. %s - %s\n", i+1, comp.Name, comp.Evidence[0])
		}
	}
	fmt.Println()

	// 3. Generate actionable fixes
	fixRecommender := services.NewFixRecommender()
	fixes := fixRecommender.RecommendFixes(
		explanation.RootCause,
		enhancedBlastRadius,
	)

	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println("STEP 3: ACTIONABLE REMEDIATION PLAN")
	fmt.Println("════════════════════════════════════════════════════════════════\n")

	fmt.Print(services.FormatActionableFix(fixes))

	// 4. Incident summary
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println("INCIDENT SUMMARY FOR STAKEHOLDERS")
	fmt.Println("════════════════════════════════════════════════════════════════\n")

	fmt.Printf("What Happened:\n%s\n\n", explanation.WhatHappened)
	fmt.Printf("Why It Happened:\n%s\n\n", explanation.WhyItHappened)
	fmt.Printf("What Broke First:\n%s\n\n", explanation.WhatBrokeFirst)

	// Comparative analysis
	fmt.Println("════════════════════════════════════════════════════════════════")
	fmt.Println("WHY THIS ROOT CAUSE IS MORE LIKELY THAN OTHERS")
	fmt.Println("════════════════════════════════════════════════════════════════\n")

	if len(explanation.AlternativeCauses) > 0 {
		topScore := explanation.RootCause.ConfidenceScore
		
		fmt.Printf("Primary Cause: %s (%d points)\n", 
			explanation.RootCause.Alert.Name, topScore)
		fmt.Println("\nAlternative Causes:")
		
		for i, alt := range explanation.AlternativeCauses {
			gap := topScore - alt.ConfidenceScore
			fmt.Printf("\n%d. %s (%d points, -%d from top)\n", 
				i+1, alt.Alert.Name, alt.ConfidenceScore, gap)
			
			// Explain the gap
			if gap >= 20 {
				fmt.Println("   ➜ Significantly less likely - missing key evidence")
			} else if gap >= 10 {
				fmt.Println("   ➜ Moderately less likely - weaker correlation")
			} else {
				fmt.Println("   ➜ Close call - could be co-causal")
			}
			
			if len(alt.Evidence) > 0 {
				fmt.Println("   Evidence for this alternative:")
				for _, ev := range alt.Evidence {
					fmt.Printf("      • %s\n", ev)
				}
			}
		}

		fmt.Println("\n🎯 CONCLUSION:")
		if topScore >= 80 {
			fmt.Println("   High confidence in primary root cause.")
			fmt.Println("   Recommendation: Focus all efforts on primary cause.")
		} else if topScore >= 60 {
			fmt.Println("   Moderate confidence in primary root cause.")
			fmt.Println("   Recommendation: Address primary cause but monitor alternatives.")
		} else {
			fmt.Println("   Multiple equally viable root causes detected.")
			fmt.Println("   Recommendation: Investigate top 2-3 causes in parallel.")
		}
	}

	fmt.Println("\n════════════════════════════════════════════════════════════════")
	fmt.Println("✅ Analysis complete. Follow the immediate fix steps above.")
	fmt.Println("════════════════════════════════════════════════════════════════")
}
