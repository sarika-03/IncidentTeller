package services

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"incident-teller/internal/domain"
)

// IncidentStory represents a narrative-style incident report
type IncidentStory struct {
	Timeline    string
	RootCause   string
	Impact      string
	Fix         IncidentFix
	Summary     string
	GeneratedAt time.Time
}

// IncidentFix contains actionable remediation steps
type IncidentFix struct {
	ImmediateActions  []string // Right now (< 5 min)
	ShortTermActions  []string // Today (< 8 hours)
	LongTermActions   []string // Prevention (ongoing)
}

// IncidentTeller converts technical incident data into human-readable stories
type IncidentTeller struct {
	comprehensiveAnalyzer *ComprehensiveIncidentAnalyzer
}

// NewIncidentTeller creates a new incident storyteller
func NewIncidentTeller() *IncidentTeller {
	return &IncidentTeller{
		comprehensiveAnalyzer: NewComprehensiveIncidentAnalyzer(),
	}
}

// TellStory converts incident alerts into a narrative story
func (it *IncidentTeller) TellStory(alerts []domain.Alert) IncidentStory {
	if len(alerts) == 0 {
		return IncidentStory{
			Summary:     "No incident detected",
			GeneratedAt: time.Now(),
		}
	}

	// Sort alerts chronologically
	sortedAlerts := make([]domain.Alert, len(alerts))
	copy(sortedAlerts, alerts)
	sort.Slice(sortedAlerts, func(i, j int) bool {
		return sortedAlerts[i].OccurredAt.Before(sortedAlerts[j].OccurredAt)
	})

	// Perform comprehensive analysis
	intelligence := it.comprehensiveAnalyzer.Analyze(sortedAlerts)

	// Generate narrative sections
	timeline := it.narrateTimeline(sortedAlerts, intelligence)
	rootCause := it.narrateRootCause(intelligence)
	impact := it.narrateImpact(intelligence)
	fix := it.narrateFixes(intelligence)
	summary := it.generateSummary(sortedAlerts, intelligence)

	return IncidentStory{
		Timeline:    timeline,
		RootCause:   rootCause,
		Impact:      impact,
		Fix:         fix,
		Summary:     summary,
		GeneratedAt: time.Now(),
	}
}

// narrateTimeline creates a cause → effect timeline narrative
func (it *IncidentTeller) narrateTimeline(
	alerts []domain.Alert,
	intelligence IncidentIntelligence,
) string {
	var narrative strings.Builder

	narrative.WriteString("Here's what happened:\n\n")

	// Group alerts by time proximity (within 2 minutes = same event cluster)
	clusters := it.clusterEvents(alerts)

	for i, cluster := range clusters {
		firstAlert := &cluster[0]
		timestamp := firstAlert.OccurredAt.Format("15:04:05")

		if i == 0 {
			// First event - the trigger
			narrative.WriteString(fmt.Sprintf("At %s, we first noticed %s on %s ",
				timestamp,
				it.describeAlert(firstAlert),
				firstAlert.Host))

			if firstAlert.Value > 90 {
				narrative.WriteString(fmt.Sprintf("hitting %.1f%%. ", firstAlert.Value))
			} else {
				narrative.WriteString(fmt.Sprintf("at %.1f%%. ", firstAlert.Value))
			}

			if len(cluster) > 1 {
				narrative.WriteString(fmt.Sprintf("Around the same time, %s also showed issues. ",
					it.listOtherAlerts(cluster[1:])))
			}
		} else {
			// Subsequent events - cascading effects
			timeSinceLast := firstAlert.OccurredAt.Sub(clusters[i-1][0].OccurredAt)

			narrative.WriteString(fmt.Sprintf("\n%s later (%s), this caused %s ",
				it.formatDuration(timeSinceLast),
				timestamp,
				it.describeAlert(firstAlert)))

			if firstAlert.ResourceType != alerts[0].ResourceType {
				narrative.WriteString("to degrade ")
			}

			narrative.WriteString(fmt.Sprintf("(%.1f%%). ", firstAlert.Value))

			if len(cluster) > 1 {
				narrative.WriteString(fmt.Sprintf("We also saw %s failing. ",
					it.listOtherAlerts(cluster[1:])))
			}
		}
	}

	// Final state
	lastAlert := alerts[len(alerts)-1]
	totalDuration := lastAlert.OccurredAt.Sub(alerts[0].OccurredAt)

	narrative.WriteString(fmt.Sprintf("\n\nThe situation fully developed over %s.",
		it.formatDuration(totalDuration)))

	return narrative.String()
}

// narrateRootCause explains the root cause in plain English
func (it *IncidentTeller) narrateRootCause(intelligence IncidentIntelligence) string {
	var narrative strings.Builder

	rc := intelligence.RootCause

	narrative.WriteString("Looking at the timeline and correlation patterns, ")

	// Confidence-based language
	if rc.ConfidenceScore >= 90 {
		narrative.WriteString("I'm highly confident that ")
	} else if rc.ConfidenceScore >= 75 {
		narrative.WriteString("the evidence strongly suggests ")
	} else if rc.ConfidenceScore >= 60 {
		narrative.WriteString("it appears ")
	} else {
		narrative.WriteString("it's possible that ")
	}

	narrative.WriteString(fmt.Sprintf("the root cause was %s ",
		it.describeAlert(rc.Alert)))

	narrative.WriteString(fmt.Sprintf("on %s. ", rc.Alert.Host))

	// Add reasoning
	if rc.IsEarliest {
		narrative.WriteString("This was the first thing to fail. ")
	}

	if rc.HasCascade {
		narrative.WriteString("After it hit critical levels, we saw a cascade effect ")
		narrative.WriteString("where other resources started degrading. ")
	}

	if rc.HasLogErrors {
		narrative.WriteString("The error logs around this time corroborate this. ")
	}

	// Alternative causes
	if len(intelligence.AlternativeCauses) > 0 {
		closeCalls := 0
		for _, alt := range intelligence.AlternativeCauses {
			if rc.ConfidenceScore-alt.ConfidenceScore < 15 {
				closeCalls++
			}
		}

		if closeCalls > 0 {
			narrative.WriteString(fmt.Sprintf("\n\nThere's also a chance that %s ",
				it.describeAlert(intelligence.AlternativeCauses[0].Alert)))
			narrative.WriteString("contributed, but the timing makes the primary cause more likely.")
		}
	}

	return narrative.String()
}

// narrateImpact describes the blast radius in human terms
func (it *IncidentTeller) narrateImpact(intelligence IncidentIntelligence) string {
	var narrative strings.Builder

	br := intelligence.BlastRadius

	// Start with simple summary
	narrative.WriteString(br.SimpleSummary)
	narrative.WriteString("\n\n")

	// Severity assessment
	if br.ImpactScore >= 80 {
		narrative.WriteString("This is a significant incident. ")
	} else if br.ImpactScore >= 60 {
		narrative.WriteString("This is a moderate incident. ")
	} else {
		narrative.WriteString("This is a relatively contained incident. ")
	}

	// Affected components
	if len(br.AffectedHosts) == 1 {
		narrative.WriteString(fmt.Sprintf("Only %s was directly affected, ",
			br.AffectedHosts[0]))
	} else {
		narrative.WriteString(fmt.Sprintf("%d hosts were affected, ",
			len(br.AffectedHosts)))
	}

	narrative.WriteString(fmt.Sprintf("with %d resource types experiencing issues. ",
		len(br.AffectedResources)))

	// Cascade info
	if br.CascadeDepth > 0 {
		narrative.WriteString(fmt.Sprintf("\n\nThe cascade went %d levels deep: ",
			br.CascadeDepth))

		// Describe the cascade chain
		directTypes := make(map[domain.ResourceType]bool)
		for _, comp := range br.DirectlyAffected {
			if comp.Type == "resource" && strings.Contains(comp.Name, " on ") {
				parts := strings.Split(comp.Name, " on ")
				if len(parts) > 0 {
					directTypes[domain.ResourceType(parts[0])] = true
				}
			}
		}

		indirectTypes := make(map[domain.ResourceType]bool)
		for _, comp := range br.IndirectlyAffected {
			if comp.Type == "resource" && strings.Contains(comp.Name, " on ") {
				parts := strings.Split(comp.Name, " on ")
				if len(parts) > 0 {
					indirectTypes[domain.ResourceType(parts[0])] = true
				}
			}
		}

		if len(directTypes) > 0 {
			narrative.WriteString(fmt.Sprintf("%v failed first, ",
				it.resourceTypesList(directTypes)))
		}

		if len(indirectTypes) > 0 {
			narrative.WriteString(fmt.Sprintf("then %v degraded as a result.",
				it.resourceTypesList(indirectTypes)))
		}
	}

	// User impact
	narrative.WriteString("\n\n")
	if br.CriticalAlerts >= 3 {
		narrative.WriteString("User-facing services were likely impacted during this time.")
	} else if br.CriticalAlerts >= 1 {
		narrative.WriteString("Some user impact is possible, though services remained partially available.")
	} else {
		narrative.WriteString("User impact was minimal - we caught this before it became user-facing.")
	}

	return narrative.String()
}

// narrateFixes provides specific, actionable remediation steps
func (it *IncidentTeller) narrateFixes(intelligence IncidentIntelligence) IncidentFix {
	fixes := intelligence.ActionableFixes

	// Rewrite in narrative style
	immediate := make([]string, 0, len(fixes.ImmediateFix))
	shortTerm := make([]string, 0, len(fixes.ShortTermFix))
	longTerm := make([]string, 0, len(fixes.LongTermFix))

	// Filter out cascade warnings and metadata, keep only actions
	for _, action := range fixes.ImmediateFix {
		if !strings.Contains(action, "CASCADE DETECTED") &&
			!strings.Contains(action, "Target host:") &&
			!strings.Contains(action, "CRITICAL:") &&
			!strings.Contains(action, "HIGH:") {
			immediate = append(immediate, action)
		}
	}

	for _, action := range fixes.ShortTermFix {
		shortTerm = append(shortTerm, action)
	}

	for _, action := range fixes.LongTermFix {
		longTerm = append(longTerm, action)
	}

	return IncidentFix{
		ImmediateActions: immediate,
		ShortTermActions: shortTerm,
		LongTermActions:  longTerm,
	}
}

// generateSummary creates a one-line incident summary
func (it *IncidentTeller) generateSummary(
	alerts []domain.Alert,
	intelligence IncidentIntelligence,
) string {
	duration := intelligence.IncidentDuration

	return fmt.Sprintf("%s on %s caused %s incident lasting %s",
		intelligence.RootCause.Alert.Name,
		intelligence.RootCause.Alert.Host,
		strings.ToLower(getSeverityLabel(intelligence.BlastRadius.ImpactScore)),
		it.formatDuration(duration))
}

// Helper methods

func (it *IncidentTeller) clusterEvents(alerts []domain.Alert) [][]domain.Alert {
	if len(alerts) == 0 {
		return nil
	}

	clusters := [][]domain.Alert{}
	currentCluster := []domain.Alert{alerts[0]}

	for i := 1; i < len(alerts); i++ {
		timeDiff := alerts[i].OccurredAt.Sub(currentCluster[0].OccurredAt)

		if timeDiff <= 2*time.Minute {
			currentCluster = append(currentCluster, alerts[i])
		} else {
			clusters = append(clusters, currentCluster)
			currentCluster = []domain.Alert{alerts[i]}
		}
	}

	clusters = append(clusters, currentCluster)
	return clusters
}

func (it *IncidentTeller) describeAlert(alert *domain.Alert) string {
	switch alert.ResourceType {
	case domain.ResourceMemory:
		return "memory pressure"
	case domain.ResourceDisk:
		return "disk space/I/O issues"
	case domain.ResourceCPU:
		return "CPU load"
	case domain.ResourceNetwork:
		return "network/latency problems"
	case domain.ResourceProcess:
		return "process failures"
	default:
		return strings.ToLower(string(alert.ResourceType)) + " issues"
	}
}

func (it *IncidentTeller) listOtherAlerts(alerts []domain.Alert) string {
	if len(alerts) == 0 {
		return ""
	}

	if len(alerts) == 1 {
		return it.describeAlert(&alerts[0])
	}

	parts := make([]string, len(alerts))
	for i, alert := range alerts {
		parts[i] = it.describeAlert(&alert)
	}

	return strings.Join(parts, " and ")
}

func (it *IncidentTeller) formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	} else if d < time.Hour {
		mins := int(d.Minutes())
		if mins == 1 {
			return "about a minute"
		}
		return fmt.Sprintf("%d minutes", mins)
	} else {
		hours := int(d.Hours())
		mins := int(d.Minutes()) - (hours * 60)
		if mins == 0 {
			return fmt.Sprintf("%d hours", hours)
		}
		return fmt.Sprintf("%d hours and %d minutes", hours, mins)
	}
}

func (it *IncidentTeller) resourceTypesList(types map[domain.ResourceType]bool) string {
	list := make([]string, 0, len(types))
	for rt := range types {
		list = append(list, strings.ToLower(string(rt)))
	}
	return strings.Join(list, ", ")
}

// FormatIncidentStory creates a formatted output for the incident story
func FormatIncidentStory(story IncidentStory) string {
	var output strings.Builder

	output.WriteString("╔════════════════════════════════════════════════════════════════╗\n")
	output.WriteString("║                    INCIDENT STORY                              ║\n")
	output.WriteString("╚════════════════════════════════════════════════════════════════╝\n\n")

	output.WriteString("📝 SUMMARY\n")
	output.WriteString("────────────────────────────────────────────────────────────────\n")
	output.WriteString(story.Summary)
	output.WriteString("\n\n")

	output.WriteString("⏱️  TIMELINE\n")
	output.WriteString("────────────────────────────────────────────────────────────────\n")
	output.WriteString(story.Timeline)
	output.WriteString("\n\n")

	output.WriteString("🎯 ROOT CAUSE\n")
	output.WriteString("────────────────────────────────────────────────────────────────\n")
	output.WriteString(story.RootCause)
	output.WriteString("\n\n")

	output.WriteString("💥 IMPACT\n")
	output.WriteString("────────────────────────────────────────────────────────────────\n")
	output.WriteString(story.Impact)
	output.WriteString("\n\n")

	output.WriteString("🔧 FIX\n")
	output.WriteString("────────────────────────────────────────────────────────────────\n\n")

	output.WriteString("IMMEDIATE (do this now):\n")
	for i, action := range story.Fix.ImmediateActions {
		output.WriteString(fmt.Sprintf("  %d. %s\n", i+1, action))
	}

	output.WriteString("\nSHORT-TERM (today):\n")
	for i, action := range story.Fix.ShortTermActions {
		output.WriteString(fmt.Sprintf("  %d. %s\n", i+1, action))
	}

	output.WriteString("\nLONG-TERM (prevention):\n")
	for i, action := range story.Fix.LongTermActions {
		output.WriteString(fmt.Sprintf("  %d. %s\n", i+1, action))
	}

	output.WriteString("\n────────────────────────────────────────────────────────────────\n")
	output.WriteString(fmt.Sprintf("Generated: %s\n", story.GeneratedAt.Format("2006-01-02 15:04:05 MST")))

	return output.String()
}
