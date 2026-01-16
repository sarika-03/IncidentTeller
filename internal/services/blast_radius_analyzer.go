package services

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"incident-teller/internal/domain"
)

// ComponentImpact represents the impact level on a component
type ComponentImpact string

const (
	ImpactDirect   ComponentImpact = "DIRECT"
	ImpactIndirect ComponentImpact = "INDIRECT"
	ImpactNone     ComponentImpact = "NONE"
)

// Component represents a system component (host, service, resource)
type Component struct {
	Name         string
	Type         string // "host", "service", "resource", "chart"
	Impact       ComponentImpact
	Evidence     []string
	AffectedAt   *time.Time
	MetricValues []float64
}

// EnhancedBlastRadiusAnalysis provides detailed impact classification
type EnhancedBlastRadiusAnalysis struct {
	DirectlyAffected   []Component
	IndirectlyAffected []Component
	Unaffected         []Component
	
	// Original fields
	AffectedHosts      []string
	AffectedResources  []domain.ResourceType
	AffectedCharts     []string
	CascadeDepth       int
	TotalAlerts        int
	CriticalAlerts     int
	Duration           time.Duration
	ImpactDescription  string
	
	// New fields
	SimpleSummary      string
	ImpactScore        int // 0-100
	RecoveryEstimate   string
}

// ActionableFix provides structured remediation guidance
type ActionableFix struct {
	ImmediateFix    []string // Actions to take right now (< 5 min)
	ShortTermFix    []string // Actions for today (< 8 hours)
	LongTermFix     []string // Prevention measures (ongoing)
	RootCauseType   domain.ResourceType
	FixComplexity   string // "Simple", "Moderate", "Complex"
	EstimatedTimeToResolve string
}

// BlastRadiusAnalyzer provides enhanced impact analysis
type BlastRadiusAnalyzer struct {
	// Known infrastructure components for comparison
	knownHosts    []string
	knownServices []string
}

// NewBlastRadiusAnalyzer creates a new enhanced analyzer
func NewBlastRadiusAnalyzer() *BlastRadiusAnalyzer {
	return &BlastRadiusAnalyzer{
		knownHosts:    []string{}, // Will be populated from infrastructure discovery
		knownServices: []string{}, // Will be populated from service registry
	}
}

// AnalyzeBlastRadius performs comprehensive impact analysis
func (b *BlastRadiusAnalyzer) AnalyzeBlastRadius(
	alerts []domain.Alert,
	rootCause RootCauseCandidate,
) EnhancedBlastRadiusAnalysis {
	
	if len(alerts) == 0 {
		return EnhancedBlastRadiusAnalysis{
			SimpleSummary: "No impact detected",
			ImpactScore:   0,
		}
	}

	// Analyze component impacts
	directComponents := b.identifyDirectlyAffected(alerts, rootCause)
	indirectComponents := b.identifyIndirectlyAffected(alerts, rootCause, directComponents)
	unaffectedComponents := b.identifyUnaffected(alerts, directComponents, indirectComponents)

	// Calculate basic metrics
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
		maxDepth = 0
	} else if len(resources) == 2 {
		maxDepth = 1
	} else {
		maxDepth = len(resources) - 1
	}

	duration := time.Duration(0)
	if len(alerts) > 0 {
		duration = alerts[len(alerts)-1].OccurredAt.Sub(alerts[0].OccurredAt)
	}

	// Calculate impact score (0-100)
	impactScore := b.calculateImpactScore(
		len(hosts),
		len(resources),
		criticalCount,
		len(alerts),
		maxDepth,
	)

	// Generate descriptions
	impactDesc := b.generateImpactDescription(len(hosts), len(resources), criticalCount)
	simpleSummary := b.generateSimpleSummary(directComponents, indirectComponents, duration)
	recoveryEstimate := b.estimateRecovery(impactScore, maxDepth, duration)

	return EnhancedBlastRadiusAnalysis{
		DirectlyAffected:   directComponents,
		IndirectlyAffected: indirectComponents,
		Unaffected:         unaffectedComponents,
		AffectedHosts:      stringKeys(hosts),
		AffectedResources:  resourceKeys(resources),
		AffectedCharts:     stringKeys(charts),
		CascadeDepth:       maxDepth,
		TotalAlerts:        len(alerts),
		CriticalAlerts:     criticalCount,
		Duration:           duration,
		ImpactDescription:  impactDesc,
		SimpleSummary:      simpleSummary,
		ImpactScore:        impactScore,
		RecoveryEstimate:   recoveryEstimate,
	}
}

// identifyDirectlyAffected finds components with primary failures
func (b *BlastRadiusAnalyzer) identifyDirectlyAffected(
	alerts []domain.Alert,
	rootCause RootCauseCandidate,
) []Component {
	components := make(map[string]*Component)

	for i := range alerts {
		alert := &alerts[i]
		
		// Only consider problem states
		if alert.Status == domain.StatusClear {
			continue
		}

		// Direct impact if:
		// 1. Same resource type as root cause
		// 2. Critical severity
		// 3. No cascading relationship
		
		isDirect := false
		evidence := []string{}

		if alert.ResourceType == rootCause.Alert.ResourceType {
			isDirect = true
			evidence = append(evidence, "Same resource type as root cause")
		}

		if alert.Status == domain.StatusCritical {
			isDirect = true
			evidence = append(evidence, "Critical severity alert")
		}

		// Host component
		hostKey := fmt.Sprintf("host:%s", alert.Host)
		if _, exists := components[hostKey]; !exists && isDirect {
			components[hostKey] = &Component{
				Name:         alert.Host,
				Type:         "host",
				Impact:       ImpactDirect,
				Evidence:     append([]string{}, evidence...),
				AffectedAt:   &alert.OccurredAt,
				MetricValues: []float64{alert.Value},
			}
		}

		// Resource component
		resourceKey := fmt.Sprintf("resource:%s:%s", alert.Host, alert.ResourceType)
		if _, exists := components[resourceKey]; !exists && isDirect {
			components[resourceKey] = &Component{
				Name:         fmt.Sprintf("%s on %s", alert.ResourceType, alert.Host),
				Type:         "resource",
				Impact:       ImpactDirect,
				Evidence:     append([]string{}, evidence...),
				AffectedAt:   &alert.OccurredAt,
				MetricValues: []float64{alert.Value},
			}
		} else if existing, exists := components[resourceKey]; exists {
			existing.MetricValues = append(existing.MetricValues, alert.Value)
		}

		// Chart component
		chartKey := fmt.Sprintf("chart:%s", alert.Chart)
		if _, exists := components[chartKey]; !exists && isDirect {
			components[chartKey] = &Component{
				Name:         alert.Chart,
				Type:         "chart",
				Impact:       ImpactDirect,
				Evidence:     append([]string{}, evidence...),
				AffectedAt:   &alert.OccurredAt,
				MetricValues: []float64{alert.Value},
			}
		}
	}

	return flattenComponents(components)
}

// identifyIndirectlyAffected finds cascading failures
func (b *BlastRadiusAnalyzer) identifyIndirectlyAffected(
	alerts []domain.Alert,
	rootCause RootCauseCandidate,
	directComponents []Component,
) []Component {
	components := make(map[string]*Component)
	directResourceTypes := make(map[domain.ResourceType]bool)
	
	// Build set of directly affected resource types
	for i := range alerts {
		if alerts[i].ResourceType == rootCause.Alert.ResourceType {
			directResourceTypes[alerts[i].ResourceType] = true
		}
	}

	for i := range alerts {
		alert := &alerts[i]
		
		if alert.Status == domain.StatusClear {
			continue
		}

		// Indirect impact if:
		// 1. Different resource type from root cause
		// 2. Occurred after root cause
		// 3. Within cascade time window
		
		isIndirect := false
		evidence := []string{}

		if alert.ResourceType != rootCause.Alert.ResourceType {
			timeDiff := alert.OccurredAt.Sub(rootCause.Alert.OccurredAt)
			if timeDiff > 0 && timeDiff <= 10*time.Minute {
				isIndirect = true
				evidence = append(evidence, 
					fmt.Sprintf("Occurred %.0fs after root cause", timeDiff.Seconds()))
				evidence = append(evidence, "Different resource type - likely cascade effect")
			}
		}

		if !isIndirect {
			continue
		}

		// Resource component
		resourceKey := fmt.Sprintf("resource:%s:%s", alert.Host, alert.ResourceType)
		if _, exists := components[resourceKey]; !exists {
			components[resourceKey] = &Component{
				Name:         fmt.Sprintf("%s on %s", alert.ResourceType, alert.Host),
				Type:         "resource",
				Impact:       ImpactIndirect,
				Evidence:     evidence,
				AffectedAt:   &alert.OccurredAt,
				MetricValues: []float64{alert.Value},
			}
		}
	}

	return flattenComponents(components)
}

// identifyUnaffected finds components that should have been affected but weren't
func (b *BlastRadiusAnalyzer) identifyUnaffected(
	alerts []domain.Alert,
	directComponents, indirectComponents []Component,
) []Component {
	// In a real system, this would query infrastructure inventory
	// and compare with affected components
	
	// For now, identify resource types that weren't affected
	affectedResources := make(map[domain.ResourceType]bool)
	for i := range alerts {
		affectedResources[alerts[i].ResourceType] = true
	}

	allResourceTypes := []domain.ResourceType{
		domain.ResourceCPU,
		domain.ResourceMemory,
		domain.ResourceDisk,
		domain.ResourceNetwork,
		domain.ResourceProcess,
	}

	unaffected := []Component{}
	for _, rt := range allResourceTypes {
		if !affectedResources[rt] {
			unaffected = append(unaffected, Component{
				Name:     string(rt),
				Type:     "resource",
				Impact:   ImpactNone,
				Evidence: []string{"No alerts detected for this resource"},
			})
		}
	}

	return unaffected
}

// calculateImpactScore computes overall impact severity (0-100)
func (b *BlastRadiusAnalyzer) calculateImpactScore(
	hosts, resources, critical, total, cascadeDepth int,
) int {
	score := 0

	// Host count (max 30 points)
	if hosts == 1 {
		score += 10
	} else if hosts <= 3 {
		score += 20
	} else {
		score += 30
	}

	// Resource diversity (max 25 points)
	score += resources * 5
	if score > 55 {
		score = 55
	}

	// Critical alerts (max 25 points)
	criticalRatio := float64(critical) / float64(total)
	score += int(criticalRatio * 25)

	// Cascade depth (max 20 points)
	score += cascadeDepth * 7
	if score > 100 {
		score = 100
	}

	return score
}

// generateImpactDescription creates human-readable impact summary
func (b *BlastRadiusAnalyzer) generateImpactDescription(hosts, resources, critical int) string {
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

// generateSimpleSummary creates a plain English blast radius summary
func (b *BlastRadiusAnalyzer) generateSimpleSummary(
	direct, indirect []Component,
	duration time.Duration,
) string {
	parts := []string{}

	// Direct impact
	directHosts := make(map[string]bool)
	directResources := make(map[string]bool)
	for _, comp := range direct {
		if comp.Type == "host" {
			directHosts[comp.Name] = true
		} else if comp.Type == "resource" {
			directResources[comp.Name] = true
		}
	}

	if len(directHosts) > 0 {
		if len(directHosts) == 1 {
			parts = append(parts, "One server was directly hit")
		} else {
			parts = append(parts, fmt.Sprintf("%d servers were directly hit", len(directHosts)))
		}
	}

	if len(directResources) > 0 {
		parts = append(parts, fmt.Sprintf("%d critical resources failed", len(directResources)))
	}

	// Indirect impact
	indirectResources := make(map[string]bool)
	for _, comp := range indirect {
		if comp.Type == "resource" {
			indirectResources[comp.Name] = true
		}
	}

	if len(indirectResources) > 0 {
		parts = append(parts, 
			fmt.Sprintf("which caused %d more resources to degrade", len(indirectResources)))
	}

	// Duration
	if duration > 0 {
		parts = append(parts, fmt.Sprintf("The incident lasted %s", formatDuration(duration)))
	}

	if len(parts) == 0 {
		return "No significant impact detected"
	}

	summary := strings.Join(parts, ", ") + "."
	
	// Capitalize first letter
	if len(summary) > 0 {
		summary = strings.ToUpper(summary[:1]) + summary[1:]
	}

	return summary
}

// estimateRecovery predicts recovery time based on impact
func (b *BlastRadiusAnalyzer) estimateRecovery(impactScore, cascadeDepth int, duration time.Duration) string {
	baseTime := 15 // minutes

	// Adjust based on impact score
	baseTime += (impactScore / 10) * 5

	// Adjust for cascade complexity
	baseTime += cascadeDepth * 10

	// Adjust based on current duration
	if duration > 30*time.Minute {
		baseTime += 30
	} else if duration > 15*time.Minute {
		baseTime += 15
	}

	if baseTime <= 30 {
		return "15-30 minutes (if addressed immediately)"
	} else if baseTime <= 60 {
		return "30-60 minutes (requires investigation)"
	} else if baseTime <= 120 {
		return "1-2 hours (complex cascading failure)"
	} else {
		return "2+ hours (major incident with extensive impact)"
	}
}

// Helper functions

func flattenComponents(m map[string]*Component) []Component {
	result := make([]Component, 0, len(m))
	for _, comp := range m {
		result = append(result, *comp)
	}
	
	// Sort by affected time (earliest first)
	sort.Slice(result, func(i, j int) bool {
		if result[i].AffectedAt == nil {
			return false
		}
		if result[j].AffectedAt == nil {
			return true
		}
		return result[i].AffectedAt.Before(*result[j].AffectedAt)
	})
	
	return result
}

func stringKeys(m map[string]bool) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d seconds", int(d.Seconds()))
	} else if d < time.Hour {
		return fmt.Sprintf("%d minutes", int(d.Minutes()))
	} else {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) - (hours * 60)
		if minutes == 0 {
			return fmt.Sprintf("%d hours", hours)
		}
		return fmt.Sprintf("%d hours %d minutes", hours, minutes)
	}
}
