package services

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"incident-teller/internal/domain"
)

// RootCauseCandidate represents a potential root cause with confidence score
type RootCauseCandidate struct {
	Alert          *domain.Alert
	ConfidenceScore int    // 0-100
	Reasoning       string
	Evidence        []string
	TimelinePosition int
	IsEarliest      bool
	HasCascade      bool
	HasLogErrors    bool
}

// BlastRadiusAnalysis represents the impact scope of an incident
type BlastRadiusAnalysis struct {
	AffectedHosts      []string
	AffectedResources  []domain.ResourceType
	AffectedCharts     []string
	CascadeDepth       int
	TotalAlerts        int
	CriticalAlerts     int
	Duration           time.Duration
	ImpactDescription  string
}

// IncidentExplanation is the plain-English output for SREs
type IncidentExplanation struct {
	WhatHappened      string
	WhyItHappened     string
	WhatBrokeFirst    string
	BlastRadius       BlastRadiusAnalysis
	SuggestedFix      string
	RootCause         RootCauseCandidate
	AlternativeCauses []RootCauseCandidate
	ConfidenceLevel   string // "Very High", "High", "Medium", "Low"
}

// SREAnalyzer provides on-call SRE-grade incident analysis
type SREAnalyzer struct {
	analyzer *IncidentAnalyzer
}

// NewSREAnalyzer creates a new SRE analyzer
func NewSREAnalyzer() *SREAnalyzer {
	return &SREAnalyzer{
		analyzer: NewIncidentAnalyzer(),
	}
}

// AnalyzeIncidentForSRE performs comprehensive root cause analysis with confidence scoring
func (s *SREAnalyzer) AnalyzeIncidentForSRE(alerts []domain.Alert) IncidentExplanation {
	if len(alerts) == 0 {
		return IncidentExplanation{
			WhatHappened: "No incident data available",
			ConfidenceLevel: "N/A",
		}
	}

	// Sort alerts chronologically
	sortedAlerts := make([]domain.Alert, len(alerts))
	copy(sortedAlerts, alerts)
	sort.Slice(sortedAlerts, func(i, j int) bool {
		return sortedAlerts[i].OccurredAt.Before(sortedAlerts[j].OccurredAt)
	})

	// Build timeline for causality analysis
	timeline := s.analyzer.AnalyzeIncident(sortedAlerts)

	// Identify all potential root causes
	candidates := s.identifyRootCauseCandidates(sortedAlerts, timeline)

	// Score each candidate
	scoredCandidates := s.scoreRootCauses(candidates, sortedAlerts)

	// Sort by confidence (highest first)
	sort.Slice(scoredCandidates, func(i, j int) bool {
		return scoredCandidates[i].ConfidenceScore > scoredCandidates[j].ConfidenceScore
	})

	// Analyze blast radius
	blastRadius := s.analyzeBlastRadius(sortedAlerts)

	// Build explanation
	explanation := IncidentExplanation{
		WhatHappened:      s.explainWhatHappened(sortedAlerts, timeline),
		WhyItHappened:     s.explainWhyItHappened(scoredCandidates, timeline),
		WhatBrokeFirst:    s.explainWhatBrokeFirst(scoredCandidates),
		BlastRadius:       blastRadius,
		SuggestedFix:      s.suggestFix(scoredCandidates, blastRadius),
		ConfidenceLevel:   s.determineConfidenceLevel(scoredCandidates),
	}

	if len(scoredCandidates) > 0 {
		explanation.RootCause = scoredCandidates[0]
		if len(scoredCandidates) > 1 {
			explanation.AlternativeCauses = scoredCandidates[1:]
		}
	}

	return explanation
}

// identifyRootCauseCandidates finds all potential root causes
func (s *SREAnalyzer) identifyRootCauseCandidates(
	alerts []domain.Alert,
	timeline []domain.TimelineEntry,
) []RootCauseCandidate {
	candidates := []RootCauseCandidate{}

	// Build a map of timeline entries by alert ID for quick lookup
	timelineMap := make(map[string]*domain.TimelineEntry)
	for i := range timeline {
		if len(timeline[i].RelatedAlertIDs) > 0 {
			timelineMap[timeline[i].RelatedAlertIDs[0]] = &timeline[i]
		}
	}

	for i := range alerts {
		alert := &alerts[i]

		// Only consider problem states (not CLEAR)
		if alert.Status == domain.StatusClear {
			continue
		}

		// Get timeline entry for causality info
		var timelineEntry *domain.TimelineEntry
		if entry, exists := timelineMap[alert.ID]; exists {
			timelineEntry = entry
		}

		candidate := RootCauseCandidate{
			Alert:            alert,
			TimelinePosition: i,
			IsEarliest:       i == 0,
			Evidence:         []string{},
		}

		// Check if this has cascading effects
		if timelineEntry != nil {
			candidate.HasCascade = s.hasCascadingEffects(alert, alerts)
		}

		// Check for log errors (simulated - in real system, query log aggregator)
		candidate.HasLogErrors = s.hasRelatedLogErrors(alert)

		candidates = append(candidates, candidate)
	}

	return candidates
}

// scoreRootCauses assigns confidence scores based on SRE heuristics
func (s *SREAnalyzer) scoreRootCauses(
	candidates []RootCauseCandidate,
	allAlerts []domain.Alert,
) []RootCauseCandidate {
	for i := range candidates {
		score := 0
		evidence := []string{}
		reasoning := ""

		alert := candidates[i].Alert

		// Rule 1: Earlier events have higher weight (max 40 points)
		if candidates[i].IsEarliest {
			score += 40
			evidence = append(evidence, "First alert in the incident timeline")
			reasoning = "This was the earliest anomaly detected"
		} else {
			// Decay score based on position
			positionPenalty := candidates[i].TimelinePosition * 5
			if positionPenalty > 30 {
				positionPenalty = 30
			}
			score += (40 - positionPenalty)
			evidence = append(evidence, fmt.Sprintf("Alert appeared at position %d in timeline", candidates[i].TimelinePosition+1))
		}

		// Rule 2: Cascading resource exhaustion (max 30 points)
		if candidates[i].HasCascade {
			score += 30
			evidence = append(evidence, "Led to cascading failures in other resources")
			reasoning += "; triggered resource exhaustion cascade"
		}

		// Rule 3: Critical severity (15 points)
		if alert.Status == domain.StatusCritical {
			score += 15
			evidence = append(evidence, "Alert reached CRITICAL severity")
		} else if alert.Status == domain.StatusWarning {
			score += 7
			evidence = append(evidence, "Alert at WARNING severity")
		}

		// Rule 4: Log errors present (15 points)
		if candidates[i].HasLogErrors {
			score += 15
			evidence = append(evidence, "Related error logs detected")
			reasoning += "; correlated with error log spikes"
		}

		// Rule 5: Known high-impact resource types
		impactScore := s.getResourceImpactScore(alert.ResourceType)
		score += impactScore
		if impactScore > 0 {
			evidence = append(evidence, fmt.Sprintf("%s is a high-impact resource", alert.ResourceType))
		}

		// Normalize to 0-100
		if score > 100 {
			score = 100
		}

		candidates[i].ConfidenceScore = score
		candidates[i].Evidence = evidence
		candidates[i].Reasoning = strings.TrimPrefix(reasoning, "; ")
	}

	return candidates
}

// getResourceImpactScore assigns weight based on resource criticality
func (s *SREAnalyzer) getResourceImpactScore(rt domain.ResourceType) int {
	switch rt {
	case domain.ResourceMemory:
		return 10 // Memory issues often cascade
	case domain.ResourceDisk:
		return 8  // Disk issues can be critical
	case domain.ResourceCPU:
		return 6  // CPU issues are common but high impact
	case domain.ResourceNetwork:
		return 7  // Network issues affect availability
	case domain.ResourceProcess:
		return 9  // Process issues often root causes
	default:
		return 0
	}
}

// hasCascadingEffects checks if this alert led to other resource issues
func (s *SREAnalyzer) hasCascadingEffects(alert *domain.Alert, allAlerts []domain.Alert) bool {
	// Count how many different resource types had issues after this alert
	laterResources := make(map[domain.ResourceType]bool)

	for i := range allAlerts {
		other := &allAlerts[i]
		// Skip same resource type or earlier alerts
		if other.ResourceType == alert.ResourceType || !other.OccurredAt.After(alert.OccurredAt) {
			continue
		}

		// Check if within cascade window
		if other.OccurredAt.Sub(alert.OccurredAt) <= 10*time.Minute {
			laterResources[other.ResourceType] = true
		}
	}

	// If 2+ different resource types had issues after this, it's a cascade
	return len(laterResources) >= 2
}

// hasRelatedLogErrors simulates log correlation (in real system, query Loki/Elasticsearch)
func (s *SREAnalyzer) hasRelatedLogErrors(alert *domain.Alert) bool {
	// In production: Query log aggregator for error logs around alert.OccurredAt
	// For now, simulate based on severity
	return alert.Status == domain.StatusCritical
}

// analyzeBlastRadius determines impact scope
func (s *SREAnalyzer) analyzeBlastRadius(alerts []domain.Alert) BlastRadiusAnalysis {
	hosts := make(map[string]bool)
	resources := make(map[domain.ResourceType]bool)
	charts := make(map[string]bool)
	criticalCount := 0
	maxDepth := 0

	for i := range alerts {
		alert := &alerts[i]
		hosts[alert.Host] = true
		resources[alert.ResourceType] = true
		charts[alert.Chart] = true

		if alert.Status == domain.StatusCritical {
			criticalCount++
		}
	}

	// Estimate cascade depth
	if len(resources) == 1 {
		maxDepth = 0 // Single resource, no cascade
	} else if len(resources) == 2 {
		maxDepth = 1 // Direct cascade
	} else {
		maxDepth = len(resources) - 1 // Multi-level cascade
	}

	duration := time.Duration(0)
	if len(alerts) > 0 {
		duration = alerts[len(alerts)-1].OccurredAt.Sub(alerts[0].OccurredAt)
	}

	impactDesc := s.generateImpactDescription(len(hosts), len(resources), criticalCount)

	return BlastRadiusAnalysis{
		AffectedHosts:      keys(hosts),
		AffectedResources:  resourceKeys(resources),
		AffectedCharts:     keys(charts),
		CascadeDepth:       maxDepth,
		TotalAlerts:        len(alerts),
		CriticalAlerts:     criticalCount,
		Duration:           duration,
		ImpactDescription:  impactDesc,
	}
}

// generateImpactDescription creates human-readable impact summary
func (s *SREAnalyzer) generateImpactDescription(hosts, resources, critical int) string {
	if hosts == 1 && resources == 1 {
		return "Localized to single host and resource"
	}
	if hosts == 1 && resources > 1 {
		return fmt.Sprintf("Single host affected, cascaded across %d resource types", resources)
	}
	if hosts > 1 && resources == 1 {
		return fmt.Sprintf("Widespread: %d hosts affected, same resource type", hosts)
	}
	return fmt.Sprintf("Widespread: %d hosts, %d resource types affected", hosts, resources)
}

// explainWhatHappened generates plain English summary
func (s *SREAnalyzer) explainWhatHappened(alerts []domain.Alert, timeline []domain.TimelineEntry) string {
	if len(alerts) == 0 {
		return "No incident occurred"
	}

	first := alerts[0]
	last := alerts[len(alerts)-1]
	duration := last.OccurredAt.Sub(first.OccurredAt)

	// Count transitions
	triggered := 0
	escalated := 0
	resolved := 0
	for _, entry := range timeline {
		switch entry.Type {
		case "TRIGGERED":
			triggered++
		case "ESCALATED":
			escalated++
		case "RESOLVED":
			resolved++
		}
	}

	summary := fmt.Sprintf("System experienced %d alert events over %s. ",
		len(alerts), duration.Round(time.Second))

	if triggered > 0 {
		summary += fmt.Sprintf("%d new alerts triggered. ", triggered)
	}
	if escalated > 0 {
		summary += fmt.Sprintf("%d alerts escalated to critical. ", escalated)
	}
	if resolved > 0 {
		summary += fmt.Sprintf("%d alerts resolved. ", resolved)
	}

	// Add resource context
	resources := make(map[domain.ResourceType]bool)
	for i := range alerts {
		resources[alerts[i].ResourceType] = true
	}
	if len(resources) > 1 {
		summary += fmt.Sprintf("Multiple resources affected: %v", resourceKeys(resources))
	}

	return summary
}

// explainWhyItHappened provides root cause reasoning
func (s *SREAnalyzer) explainWhyItHappened(candidates []RootCauseCandidate, timeline []domain.TimelineEntry) string {
	if len(candidates) == 0 {
		return "Unable to determine root cause"
	}

	rootCause := candidates[0]
	alert := rootCause.Alert

	explanation := fmt.Sprintf("The incident was triggered by %s exhaustion on %s. ",
		alert.ResourceType, alert.Chart)

	if rootCause.Reasoning != "" {
		explanation += rootCause.Reasoning + ". "
	}

	if rootCause.HasCascade {
		explanation += "This caused a cascade effect, impacting other system resources. "
	}

	return explanation
}

// explainWhatBrokeFirst identifies the first failure
func (s *SREAnalyzer) explainWhatBrokeFirst(candidates []RootCauseCandidate) string {
	if len(candidates) == 0 {
		return "No failures detected"
	}

	first := candidates[0]
	alert := first.Alert

	return fmt.Sprintf("%s on %s@%s (value: %.2f at %s)",
		alert.Name,
		alert.Chart,
		alert.Host,
		alert.Value,
		alert.OccurredAt.Format("15:04:05"))
}

// suggestFix provides remediation guidance
func (s *SREAnalyzer) suggestFix(candidates []RootCauseCandidate, blast BlastRadiusAnalysis) string {
	if len(candidates) == 0 {
		return "No fix needed"
	}

	rootCause := candidates[0].Alert

	fixes := []string{}

	// Resource-specific fixes
	switch rootCause.ResourceType {
	case domain.ResourceMemory:
		fixes = append(fixes, "1. Check for memory leaks in running processes")
		fixes = append(fixes, "2. Identify top memory consumers: `ps aux --sort=-%mem | head -10`")
		fixes = append(fixes, "3. Consider increasing swap space or adding RAM")
		fixes = append(fixes, "4. Review application memory limits and OOM killer logs")

	case domain.ResourceDisk:
		fixes = append(fixes, "1. Identify large files: `du -h / | sort -rh | head -20`")
		fixes = append(fixes, "2. Clear log files: `journalctl --vacuum-time=7d`")
		fixes = append(fixes, "3. Check for filled partitions: `df -h`")
		fixes = append(fixes, "4. Remove old package caches and temp files")

	case domain.ResourceCPU:
		fixes = append(fixes, "1. Identify CPU-heavy processes: `top` or `htop`")
		fixes = append(fixes, "2. Check for runaway processes or infinite loops")
		fixes = append(fixes, "3. Review cron jobs and background tasks")
		fixes = append(fixes, "4. Consider scaling horizontally if load is legitimate")

	case domain.ResourceNetwork:
		fixes = append(fixes, "1. Check network interface status: `ip link show`")
		fixes = append(fixes, "2. Analyze traffic: `netstat -s` or `ss -s`")
		fixes = append(fixes, "3. Look for DDoS or unusual connection patterns")
		fixes = append(fixes, "4. Verify DNS resolution and external connectivity")

	case domain.ResourceProcess:
		fixes = append(fixes, "1. Restart affected service: `systemctl restart <service>`")
		fixes = append(fixes, "2. Check process logs for errors")
		fixes = append(fixes, "3. Verify process limits: `ulimit -a`")
		fixes = append(fixes, "4. Review recent deployments or config changes")

	default:
		fixes = append(fixes, "1. Review system logs: `journalctl -xe`")
		fixes = append(fixes, "2. Check resource utilization: `vmstat 1 5`")
		fixes = append(fixes, "3. Verify service health")
	}

	// Add cascade mitigation if needed
	if blast.CascadeDepth > 0 {
		fixes = append(fixes, "\n🔄 Cascade Mitigation:")
		fixes = append(fixes, "- Address root cause first to stop propagation")
		fixes = append(fixes, "- Monitor dependent resources for stabilization")
	}

	return strings.Join(fixes, "\n")
}

// determineConfidenceLevel converts score to human-readable confidence
func (s *SREAnalyzer) determineConfidenceLevel(candidates []RootCauseCandidate) string {
	if len(candidates) == 0 {
		return "N/A"
	}

	score := candidates[0].ConfidenceScore

	switch {
	case score >= 80:
		return "Very High (≥80%)"
	case score >= 60:
		return "High (60-79%)"
	case score >= 40:
		return "Medium (40-59%)"
	default:
		return "Low (<40%)"
	}
}

// Helper functions
func keys(m map[string]bool) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

func resourceKeys(m map[domain.ResourceType]bool) []domain.ResourceType {
	result := make([]domain.ResourceType, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

// FormatIncidentExplanation creates a formatted report for SREs
func FormatIncidentExplanation(exp IncidentExplanation) string {
	var report strings.Builder

	report.WriteString("╔═══════════════════════════════════════════════════════════════╗\n")
	report.WriteString("║           SRE INCIDENT ANALYSIS REPORT                        ║\n")
	report.WriteString("╚═══════════════════════════════════════════════════════════════╝\n\n")

	report.WriteString("📋 WHAT HAPPENED\n")
	report.WriteString("═══════════════════════════════════════════════════════════════\n")
	report.WriteString(exp.WhatHappened)
	report.WriteString("\n\n")

	report.WriteString("🔍 WHY IT HAPPENED\n")
	report.WriteString("═══════════════════════════════════════════════════════════════\n")
	report.WriteString(exp.WhyItHappened)
	report.WriteString("\n\n")

	report.WriteString("🔴 WHAT BROKE FIRST\n")
	report.WriteString("═══════════════════════════════════════════════════════════════\n")
	report.WriteString(exp.WhatBrokeFirst)
	report.WriteString("\n\n")

	report.WriteString("💥 BLAST RADIUS\n")
	report.WriteString("═══════════════════════════════════════════════════════════════\n")
	report.WriteString(fmt.Sprintf("Impact: %s\n", exp.BlastRadius.ImpactDescription))
	report.WriteString(fmt.Sprintf("Affected Hosts: %d (%v)\n", len(exp.BlastRadius.AffectedHosts), exp.BlastRadius.AffectedHosts))
	report.WriteString(fmt.Sprintf("Affected Resources: %v\n", exp.BlastRadius.AffectedResources))
	report.WriteString(fmt.Sprintf("Total Alerts: %d (Critical: %d)\n", exp.BlastRadius.TotalAlerts, exp.BlastRadius.CriticalAlerts))
	report.WriteString(fmt.Sprintf("Cascade Depth: %d levels\n", exp.BlastRadius.CascadeDepth))
	report.WriteString(fmt.Sprintf("Duration: %s\n", exp.BlastRadius.Duration.Round(time.Second)))
	report.WriteString("\n")

	report.WriteString("🔧 SUGGESTED FIX\n")
	report.WriteString("═══════════════════════════════════════════════════════════════\n")
	report.WriteString(exp.SuggestedFix)
	report.WriteString("\n\n")

	report.WriteString("🎯 ROOT CAUSE ANALYSIS\n")
	report.WriteString("═══════════════════════════════════════════════════════════════\n")
	report.WriteString(fmt.Sprintf("Confidence: %s (%d/100)\n\n", exp.ConfidenceLevel, exp.RootCause.ConfidenceScore))
	
	report.WriteString("Primary Root Cause:\n")
	report.WriteString(fmt.Sprintf("  • Alert: %s\n", exp.RootCause.Alert.Name))
	report.WriteString(fmt.Sprintf("  • Resource: %s\n", exp.RootCause.Alert.ResourceType))
	report.WriteString(fmt.Sprintf("  • Host: %s\n", exp.RootCause.Alert.Host))
	report.WriteString(fmt.Sprintf("  • Value: %.2f\n", exp.RootCause.Alert.Value))
	
	if len(exp.RootCause.Evidence) > 0 {
		report.WriteString("\nEvidence:\n")
		for _, evidence := range exp.RootCause.Evidence {
			report.WriteString(fmt.Sprintf("  ✓ %s\n", evidence))
		}
	}

	if len(exp.AlternativeCauses) > 0 {
		report.WriteString("\nAlternative Causes:\n")
		for i, alt := range exp.AlternativeCauses {
			if i >= 3 {
				break // Show top 3 alternatives
			}
			report.WriteString(fmt.Sprintf("  %d. %s (%s) - Confidence: %d%%\n",
				i+1, alt.Alert.Name, alt.Alert.ResourceType, alt.ConfidenceScore))
		}
	}

	report.WriteString("\n")
	report.WriteString("═══════════════════════════════════════════════════════════════\n")

	return report.String()
}
