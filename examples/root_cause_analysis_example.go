package main

import (
	"fmt"
	"time"

	"incident-teller/internal/domain"
	"incident-teller/internal/services"
)

func main() {
	// Create SRE Analyzer
	sreAnalyzer := services.NewSREAnalyzer()

	// Example: Simulate an incident timeline with multiple events
	alerts := []domain.Alert{
		{
			ID:          "alert-001",
			Name:        "memory_usage_high",
			Status:      domain.StatusWarning,
			OldStatus:   domain.StatusClear,
			ResourceType: domain.ResourceMemory,
			Chart:       "system.ram",
			Host:        "prod-server-01",
			Value:       85.3,
			OccurredAt:  time.Now().Add(-10 * time.Minute),
			Description: "Memory usage exceeded 85%",
		},
		{
			ID:          "alert-002",
			Name:        "swap_usage_critical",
			Status:      domain.StatusCritical,
			OldStatus:   domain.StatusWarning,
			ResourceType: domain.ResourceDisk,
			Chart:       "system.swap",
			Host:        "prod-server-01",
			Value:       95.7,
			OccurredAt:  time.Now().Add(-7 * time.Minute),
			Description: "Swap usage critical - disk thrashing",
		},
		{
			ID:          "alert-003",
			Name:        "cpu_iowait_high",
			Status:      domain.StatusCritical,
			OldStatus:   domain.StatusClear,
			ResourceType: domain.ResourceCPU,
			Chart:       "system.cpu",
			Host:        "prod-server-01",
			Value:       78.2,
			OccurredAt:  time.Now().Add(-5 * time.Minute),
			Description: "High CPU iowait due to disk saturation",
		},
		{
			ID:          "alert-004",
			Name:        "response_time_degraded",
			Status:      domain.StatusWarning,
			OldStatus:   domain.StatusClear,
			ResourceType: domain.ResourceNetwork,
			Chart:       "app.response_time",
			Host:        "prod-server-01",
			Value:       1250.0,
			OccurredAt:  time.Now().Add(-3 * time.Minute),
			Description: "Application response time increased",
		},
	}

	// Perform SRE-grade incident analysis
	explanation := sreAnalyzer.AnalyzeIncidentForSRE(alerts)

	// Print formatted report
	fmt.Println(services.FormatIncidentExplanation(explanation))

	// Access specific fields programmatically
	fmt.Println("\n\n🎯 PROGRAMMATIC ACCESS:")
	fmt.Println("========================================")
	fmt.Printf("Root Cause: %s\n", explanation.RootCause.Alert.Name)
	fmt.Printf("Confidence Score: %d/100\n", explanation.RootCause.ConfidenceScore)
	fmt.Printf("Confidence Level: %s\n", explanation.ConfidenceLevel)
	
	fmt.Println("\n📊 Why this is the most likely cause:")
	fmt.Printf("Reasoning: %s\n", explanation.RootCause.Reasoning)
	
	fmt.Println("\nEvidence:")
	for _, evidence := range explanation.RootCause.Evidence {
		fmt.Printf("  • %s\n", evidence)
	}

	if len(explanation.AlternativeCauses) > 0 {
		fmt.Println("\n🔄 Alternative root causes considered:")
		for i, alt := range explanation.AlternativeCauses {
			fmt.Printf("  %d. %s (Confidence: %d/100)\n", 
				i+1, alt.Alert.Name, alt.ConfidenceScore)
		}
	}

	// Show comparison between candidates
	fmt.Println("\n🔍 COMPARATIVE ANALYSIS:")
	fmt.Println("========================================")
	fmt.Printf("Top candidate score: %d\n", explanation.RootCause.ConfidenceScore)
	if len(explanation.AlternativeCauses) > 0 {
		secondScore := explanation.AlternativeCauses[0].ConfidenceScore
		gap := explanation.RootCause.ConfidenceScore - secondScore
		fmt.Printf("Second candidate score: %d\n", secondScore)
		fmt.Printf("Confidence gap: %d points\n", gap)
		
		if gap >= 30 {
			fmt.Println("✓ Strong differentiation - high confidence in primary cause")
		} else if gap >= 15 {
			fmt.Println("⚠ Moderate differentiation - consider multiple causes")
		} else {
			fmt.Println("⚠ Weak differentiation - multiple equally likely causes")
		}
	}
}
