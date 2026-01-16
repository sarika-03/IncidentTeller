package services

import (
	"fmt"
	"time"

	"incident-teller/internal/domain"
)

// IncidentIntelligence provides the complete SRE analysis package
type IncidentIntelligence struct {
	// Root cause analysis
	RootCause           RootCauseCandidate
	AlternativeCauses   []RootCauseCandidate
	ConfidenceLevel     string
	
	// Blast radius
	BlastRadius         EnhancedBlastRadiusAnalysis
	
	// Remediation
	ActionableFixes     ActionableFix
	
	// Narrative
	WhatHappened        string
	WhyItHappened       string
	WhatBrokeFirst      string
	
	// Metadata
	AnalyzedAt          time.Time
	TotalAlerts         int
	IncidentDuration    time.Duration
}

// ComprehensiveIncidentAnalyzer orchestrates all analysis components
type ComprehensiveIncidentAnalyzer struct {
	sreAnalyzer         *SREAnalyzer
	blastRadiusAnalyzer *BlastRadiusAnalyzer
	fixRecommender      *FixRecommender
}

// NewComprehensiveIncidentAnalyzer creates the complete analyzer
func NewComprehensiveIncidentAnalyzer() *ComprehensiveIncidentAnalyzer {
	return &ComprehensiveIncidentAnalyzer{
		sreAnalyzer:         NewSREAnalyzer(),
		blastRadiusAnalyzer: NewBlastRadiusAnalyzer(),
		fixRecommender:      NewFixRecommender(),
	}
}

// Analyze performs complete incident analysis and returns intelligence package
func (c *ComprehensiveIncidentAnalyzer) Analyze(alerts []domain.Alert) IncidentIntelligence {
	startTime := time.Now()
	
	// Step 1: Root cause analysis with confidence scoring
	explanation := c.sreAnalyzer.AnalyzeIncidentForSRE(alerts)
	
	// Step 2: Enhanced blast radius analysis
	blastRadius := c.blastRadiusAnalyzer.AnalyzeBlastRadius(
		alerts,
		explanation.RootCause,
	)
	
	// Step 3: Actionable fix recommendations
	fixes := c.fixRecommender.RecommendFixes(
		explanation.RootCause,
		blastRadius,
	)
	
	// Calculate incident duration
	var duration time.Duration
	if len(alerts) > 0 {
		duration = alerts[len(alerts)-1].OccurredAt.Sub(alerts[0].OccurredAt)
	}
	
	return IncidentIntelligence{
		RootCause:         explanation.RootCause,
		AlternativeCauses: explanation.AlternativeCauses,
		ConfidenceLevel:   explanation.ConfidenceLevel,
		BlastRadius:       blastRadius,
		ActionableFixes:   fixes,
		WhatHappened:      explanation.WhatHappened,
		WhyItHappened:     explanation.WhyItHappened,
		WhatBrokeFirst:    explanation.WhatBrokeFirst,
		AnalyzedAt:        startTime,
		TotalAlerts:       len(alerts),
		IncidentDuration:  duration,
	}
}

// GenerateExecutiveSummary creates a concise summary for leadership
func (c *ComprehensiveIncidentAnalyzer) GenerateExecutiveSummary(
	intelligence IncidentIntelligence,
) string {
	summary := fmt.Sprintf(`
╔════════════════════════════════════════════════════════════════╗
║              EXECUTIVE INCIDENT SUMMARY                        ║
╚════════════════════════════════════════════════════════════════╝

📊 INCIDENT OVERVIEW
────────────────────────────────────────────────────────────────
Duration:          %s
Total Alerts:      %d (%d critical)
Impact Score:      %d/100 (%s)
Recovery Time:     %s

🎯 ROOT CAUSE (Confidence: %s)
────────────────────────────────────────────────────────────────
%s on %s
Value: %.2f
Time: %s

💥 BUSINESS IMPACT
────────────────────────────────────────────────────────────────
%s

Affected:
  • %d hosts
  • %d resource types
  • Cascade depth: %d levels

🔧 STATUS & NEXT STEPS
────────────────────────────────────────────────────────────────
Fix Complexity:    %s
Est. Resolution:   %s

Immediate Actions Required:
`,
		intelligence.IncidentDuration.Round(time.Second),
		intelligence.TotalAlerts,
		intelligence.BlastRadius.CriticalAlerts,
		intelligence.BlastRadius.ImpactScore,
		getSeverityLabel(intelligence.BlastRadius.ImpactScore),
		intelligence.BlastRadius.RecoveryEstimate,
		intelligence.ConfidenceLevel,
		intelligence.RootCause.Alert.Name,
		intelligence.RootCause.Alert.Host,
		intelligence.RootCause.Alert.Value,
		intelligence.RootCause.Alert.OccurredAt.Format("15:04:05 MST"),
		intelligence.BlastRadius.SimpleSummary,
		len(intelligence.BlastRadius.AffectedHosts),
		len(intelligence.BlastRadius.AffectedResources),
		intelligence.BlastRadius.CascadeDepth,
		intelligence.ActionableFixes.FixComplexity,
		intelligence.ActionableFixes.EstimatedTimeToResolve,
	)
	
	// Add top 3 immediate actions
	for i, action := range intelligence.ActionableFixes.ImmediateFix {
		if i >= 3 {
			break
		}
		summary += fmt.Sprintf("  %d. %s\n", i+1, action)
	}
	
	summary += "\n════════════════════════════════════════════════════════════════\n"
	
	return summary
}

// GenerateTechnicalReport creates a detailed report for on-call engineers
func (c *ComprehensiveIncidentAnalyzer) GenerateTechnicalReport(
	intelligence IncidentIntelligence,
) string {
	var report string
	
	report += "╔════════════════════════════════════════════════════════════════╗\n"
	report += "║           TECHNICAL INCIDENT ANALYSIS REPORT                   ║\n"
	report += "╚════════════════════════════════════════════════════════════════╝\n\n"
	
	// Section 1: Timeline
	report += "📅 INCIDENT TIMELINE\n"
	report += "════════════════════════════════════════════════════════════════\n"
	report += fmt.Sprintf("Start:    %s\n", intelligence.RootCause.Alert.OccurredAt.Format(time.RFC3339))
	report += fmt.Sprintf("Duration: %s\n", intelligence.IncidentDuration.Round(time.Second))
	report += fmt.Sprintf("Analyzed: %s\n\n", intelligence.AnalyzedAt.Format(time.RFC3339))
	
	// Section 2: What happened
	report += "📋 WHAT HAPPENED\n"
	report += "════════════════════════════════════════════════════════════════\n"
	report += intelligence.WhatHappened + "\n\n"
	
	// Section 3: Root cause with alternatives
	report += "🎯 ROOT CAUSE ANALYSIS\n"
	report += "════════════════════════════════════════════════════════════════\n"
	report += fmt.Sprintf("Primary: %s\n", intelligence.RootCause.Alert.Name)
	report += fmt.Sprintf("Confidence: %d/100 (%s)\n", 
		intelligence.RootCause.ConfidenceScore,
		intelligence.ConfidenceLevel)
	report += fmt.Sprintf("Reasoning: %s\n\n", intelligence.RootCause.Reasoning)
	
	if len(intelligence.RootCause.Evidence) > 0 {
		report += "Evidence:\n"
		for _, ev := range intelligence.RootCause.Evidence {
			report += fmt.Sprintf("  ✓ %s\n", ev)
		}
		report += "\n"
	}
	
	if len(intelligence.AlternativeCauses) > 0 {
		report += "Alternative Causes Evaluated:\n"
		for i, alt := range intelligence.AlternativeCauses {
			if i >= 3 {
				break
			}
			gap := intelligence.RootCause.ConfidenceScore - alt.ConfidenceScore
			report += fmt.Sprintf("  %d. %s (Confidence: %d, Gap: -%d)\n", 
				i+1, alt.Alert.Name, alt.ConfidenceScore, gap)
		}
		report += "\n"
	}
	
	// Section 4: Blast radius
	report += "💥 BLAST RADIUS\n"
	report += "════════════════════════════════════════════════════════════════\n"
	report += fmt.Sprintf("Impact Score: %d/100\n", intelligence.BlastRadius.ImpactScore)
	report += fmt.Sprintf("Summary: %s\n\n", intelligence.BlastRadius.SimpleSummary)
	
	report += fmt.Sprintf("Direct Impact:\n")
	report += fmt.Sprintf("  • %d components\n", len(intelligence.BlastRadius.DirectlyAffected))
	
	report += fmt.Sprintf("Indirect Impact (Cascade):\n")
	report += fmt.Sprintf("  • %d components (%d cascade levels)\n\n", 
		len(intelligence.BlastRadius.IndirectlyAffected),
		intelligence.BlastRadius.CascadeDepth)
	
	// Section 5: Remediation
	report += FormatActionableFix(intelligence.ActionableFixes)
	
	return report
}

// GenerateSlackMessage creates a Slack-formatted incident notification
func (c *ComprehensiveIncidentAnalyzer) GenerateSlackMessage(
	intelligence IncidentIntelligence,
) string {
	severity := getSeverityEmoji(intelligence.BlastRadius.ImpactScore)
	
	msg := fmt.Sprintf(`%s *INCIDENT ALERT*

*Root Cause:* %s (Confidence: %d%%)
*Host:* %s
*Impact:* %s (%d/100)
*Duration:* %s

*What Happened:*
%s

*Immediate Actions:*
`,
		severity,
		intelligence.RootCause.Alert.Name,
		intelligence.RootCause.ConfidenceScore,
		intelligence.RootCause.Alert.Host,
		getSeverityLabel(intelligence.BlastRadius.ImpactScore),
		intelligence.BlastRadius.ImpactScore,
		intelligence.IncidentDuration.Round(time.Second),
		intelligence.BlastRadius.SimpleSummary,
	)
	
	for i, action := range intelligence.ActionableFixes.ImmediateFix {
		if i >= 3 {
			break
		}
		msg += fmt.Sprintf("%d. %s\n", i+1, action)
	}
	
	msg += fmt.Sprintf("\n*Est. Time to Resolve:* %s", 
		intelligence.ActionableFixes.EstimatedTimeToResolve)
	
	return msg
}

// Helper functions

func getSeverityLabel(score int) string {
	switch {
	case score >= 80:
		return "CRITICAL"
	case score >= 60:
		return "HIGH"
	case score >= 40:
		return "MEDIUM"
	default:
		return "LOW"
	}
}

func getSeverityEmoji(score int) string {
	switch {
	case score >= 80:
		return "🚨"
	case score >= 60:
		return "⚠️"
	case score >= 40:
		return "⚡"
	default:
		return "ℹ️"
	}
}
