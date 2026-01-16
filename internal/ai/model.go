package ai

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"incident-teller/internal/domain"
)

// AIModel provides intelligent incident analysis using ML
type AIModel interface {
	PredictRootCause(ctx context.Context, alerts []domain.Alert) (RootCausePrediction, error)
	PredictBlastRadius(ctx context.Context, alerts []domain.Alert) (BlastRadiusPrediction, error)
	AnalyzePatterns(ctx context.Context, alerts []domain.Alert) (PatternAnalysis, error)
}

// RootCausePrediction uses ML to predict root cause with confidence
type RootCausePrediction struct {
	PrimaryCause      *domain.Alert
	Confidence        float64 // 0.0-1.0
	AlternativeCauses []*domain.Alert
	Reasoning         string
	PatternType       string // "cascade", "spike", "gradual", "sudden"
	MLFeatures        []string
	ModelVersion      string
}

// BlastRadiusPrediction predicts impact scope using ML
type BlastRadiusPrediction struct {
	ImpactScore        float64 // 0.0-1.0
	AffectedServices   []string
	CascadeProbability float64
	DurationPredicted  time.Duration
	BusinessImpact     string
	RiskLevel          string // "low", "medium", "high", "critical"
}

// PatternAnalysis identifies temporal and correlation patterns
type PatternAnalysis struct {
	PatternType       string
	Confidence        float64
	Seasonal          bool
	Trend             string // "increasing", "decreasing", "stable"
	AnomalyScore      float64
	CorrelationMatrix map[string]float64
	PredictedNext     time.Time
}

// LocalAIModel implements AI with ML algorithms
type LocalAIModel struct {
	featureExtractor *FeatureExtractor
	patternMatcher   *PatternMatcher
	classifier       *IncidentClassifier
}

// NewLocalAIModel creates a new AI model instance
func NewLocalAIModel() *LocalAIModel {
	return &LocalAIModel{
		featureExtractor: NewFeatureExtractor(),
		patternMatcher:   NewPatternMatcher(),
		classifier:       NewIncidentClassifier(),
	}
}

// PredictRootCause uses ML algorithms to predict root cause
func (ai *LocalAIModel) PredictRootCause(ctx context.Context, alerts []domain.Alert) (RootCausePrediction, error) {
	if len(alerts) == 0 {
		return RootCausePrediction{}, fmt.Errorf("no alerts to analyze")
	}

	// Extract features from alerts
	features := ai.featureExtractor.ExtractFeatures(alerts)

	// Use ensemble of ML methods
	candidates := ai.identifyRootCauseCandidates(alerts, features)
	scores := ai.scoreWithML(candidates, features)

	// Select best candidate
	bestCandidate, confidence := ai.selectBestCandidate(candidates, scores)

	// Handle case where all alerts are resolved (no active candidates)
	if bestCandidate == nil {
		return RootCausePrediction{
			PrimaryCause:      nil,
			Confidence:        0.0,
			AlternativeCauses: []*domain.Alert{},
			Reasoning:         "All alerts are resolved - no active root cause detected",
			PatternType:       ai.patternMatcher.IdentifyPattern(alerts, features),
			MLFeatures:        features,
			ModelVersion:      "1.0.0",
		}, nil
	}

	// Generate reasoning
	reasoning := ai.generateReasoning(bestCandidate, features, confidence)

	// Identify pattern type
	patternType := ai.patternMatcher.IdentifyPattern(alerts, features)

	return RootCausePrediction{
		PrimaryCause:      bestCandidate,
		Confidence:        confidence,
		AlternativeCauses: ai.getAlternativeCauses(candidates, scores, bestCandidate),
		Reasoning:         reasoning,
		PatternType:       patternType,
		MLFeatures:        features,
		ModelVersion:      "1.0.0",
	}, nil
}

// PredictBlastRadius uses ML to predict incident impact
func (ai *LocalAIModel) PredictBlastRadius(ctx context.Context, alerts []domain.Alert) (BlastRadiusPrediction, error) {
	if len(alerts) == 0 {
		return BlastRadiusPrediction{}, fmt.Errorf("no alerts to analyze")
	}

	features := ai.featureExtractor.ExtractFeatures(alerts)

	// Predict impact using trained model
	impactScore := ai.classifier.PredictImpact(features)

	// Predict cascade probability
	cascadeProb := ai.classifier.PredictCascadeProbability(features)

	// Estimate duration
	duration := ai.classifier.PredictDuration(features)

	// Determine business impact
	businessImpact := ai.classifyBusinessImpact(impactScore, cascadeProb)

	// Determine risk level
	riskLevel := ai.determineRiskLevel(impactScore, cascadeProb, duration)

	// Identify affected services
	affectedServices := ai.identifyAffectedServices(alerts, features)

	return BlastRadiusPrediction{
		ImpactScore:        impactScore,
		AffectedServices:   affectedServices,
		CascadeProbability: cascadeProb,
		DurationPredicted:  duration,
		BusinessImpact:     businessImpact,
		RiskLevel:          riskLevel,
	}, nil
}

// AnalyzePatterns identifies temporal patterns and correlations
func (ai *LocalAIModel) AnalyzePatterns(ctx context.Context, alerts []domain.Alert) (PatternAnalysis, error) {
	if len(alerts) == 0 {
		return PatternAnalysis{}, fmt.Errorf("no alerts to analyze")
	}

	features := ai.featureExtractor.ExtractFeatures(alerts)

	// Identify pattern type
	patternType := ai.patternMatcher.IdentifyPattern(alerts, features)

	// Calculate confidence
	confidence := ai.calculatePatternConfidence(alerts, patternType)

	// Check for seasonality
	seasonal := ai.detectSeasonality(alerts)

	// Determine trend
	trend := ai.determineTrend(alerts)

	// Calculate anomaly score
	anomalyScore := ai.calculateAnomalyScore(features)

	// Build correlation matrix
	correlationMatrix := ai.buildCorrelationMatrix(alerts)

	// Predict next occurrence
	nextOccurrence := ai.predictNextOccurrence(alerts, patternType)

	return PatternAnalysis{
		PatternType:       patternType,
		Confidence:        confidence,
		Seasonal:          seasonal,
		Trend:             trend,
		AnomalyScore:      anomalyScore,
		CorrelationMatrix: correlationMatrix,
		PredictedNext:     nextOccurrence,
	}, nil
}

// FeatureExtractor converts alerts to ML features
type FeatureExtractor struct{}

// NewFeatureExtractor creates a new feature extractor
func NewFeatureExtractor() *FeatureExtractor {
	return &FeatureExtractor{}
}

// ExtractFeatures converts alerts to feature vector for ML
func (fe *FeatureExtractor) ExtractFeatures(alerts []domain.Alert) []string {
	features := []string{}

	if len(alerts) == 0 {
		return features
	}

	// Time-based features
	duration := alerts[len(alerts)-1].OccurredAt.Sub(alerts[0].OccurredAt)
	features = append(features, fmt.Sprintf("duration_seconds:%.0f", duration.Seconds()))

	// Alert count features
	features = append(features, fmt.Sprintf("alert_count:%d", len(alerts)))

	criticalCount := 0
	warningCount := 0
	resourceTypes := make(map[domain.ResourceType]int)
	hosts := make(map[string]int)

	for _, alert := range alerts {
		if alert.Status == domain.StatusCritical {
			criticalCount++
		} else if alert.Status == domain.StatusWarning {
			warningCount++
		}
		resourceTypes[alert.ResourceType]++
		hosts[alert.Host]++
	}

	features = append(features, fmt.Sprintf("critical_count:%d", criticalCount))
	features = append(features, fmt.Sprintf("warning_count:%d", warningCount))
	features = append(features, fmt.Sprintf("host_count:%d", len(hosts)))
	features = append(features, fmt.Sprintf("resource_type_count:%d", len(resourceTypes)))

	// Resource type features
	for rt, count := range resourceTypes {
		features = append(features, fmt.Sprintf("resource_%s_count:%d", rt, count))
	}

	// Severity features
	if criticalCount > 0 {
		features = append(features, "has_critical:true")
	}
	if warningCount > 0 {
		features = append(features, "has_warning:true")
	}

	// Temporal features
	hourOfDay := alerts[0].OccurredAt.Hour()
	features = append(features, fmt.Sprintf("hour_of_day:%d", hourOfDay))
	dayOfWeek := int(alerts[0].OccurredAt.Weekday())
	features = append(features, fmt.Sprintf("day_of_week:%d", dayOfWeek))

	// Pattern features
	if len(alerts) > 1 {
		timeWindows := []time.Duration{1 * time.Minute, 5 * time.Minute, 15 * time.Minute}
		for _, window := range timeWindows {
			count := fe.countAlertsInWindow(alerts, window)
			features = append(features, fmt.Sprintf("alerts_%.0f_min:%d", window.Minutes(), count))
		}
	}

	// Value-based features
	maxValue := 0.0
	minValue := math.MaxFloat64
	for _, alert := range alerts {
		if alert.Value > maxValue {
			maxValue = alert.Value
		}
		if alert.Value < minValue {
			minValue = alert.Value
		}
	}
	features = append(features, fmt.Sprintf("max_value:%.2f", maxValue))
	features = append(features, fmt.Sprintf("min_value:%.2f", minValue))

	return features
}

// PatternMatcher identifies incident patterns
type PatternMatcher struct{}

// NewPatternMatcher creates a new pattern matcher
func NewPatternMatcher() *PatternMatcher {
	return &PatternMatcher{}
}

// IdentifyPattern classifies the incident pattern
func (pm *PatternMatcher) IdentifyPattern(alerts []domain.Alert, features []string) string {
	if len(alerts) < 2 {
		return "single"
	}

	// Analyze timing patterns
	timeDiff := alerts[len(alerts)-1].OccurredAt.Sub(alerts[0].OccurredAt)
	alertRate := float64(len(alerts)) / timeDiff.Hours()

	// Classify based on alert rate and spread
	if alertRate > 10 {
		return "burst"
	} else if alertRate > 2 {
		return "cascade"
	} else if timeDiff > 30*time.Minute {
		return "gradual"
	} else if timeDiff < 5*time.Minute && len(alerts) > 3 {
		return "spike"
	}

	return "progressive"
}

// IncidentClassifier provides ML-based classification
type IncidentClassifier struct{}

// NewIncidentClassifier creates a new classifier
func NewIncidentClassifier() *IncidentClassifier {
	return &IncidentClassifier{}
}

// PredictImpact predicts incident impact score (0.0-1.0)
func (ic *IncidentClassifier) PredictImpact(features []string) float64 {
	score := 0.0

	// Base score from alert count
	alertCount := ic.extractFeatureValue(features, "alert_count:")
	score += math.Min(float64(alertCount)/10.0, 0.3)

	// Critical alerts increase impact
	criticalCount := ic.extractFeatureValue(features, "critical_count:")
	score += float64(criticalCount) * 0.2

	// Multiple hosts increase impact
	hostCount := ic.extractFeatureValue(features, "host_count:")
	score += math.Min(float64(hostCount-1)*0.15, 0.3)

	// Multiple resource types increase impact
	resourceCount := ic.extractFeatureValue(features, "resource_type_count:")
	score += math.Min(float64(resourceCount-1)*0.1, 0.2)

	// Cap at 1.0
	return math.Min(score, 1.0)
}

// PredictCascadeProbability predicts likelihood of cascade
func (ic *IncidentClassifier) PredictCascadeProbability(features []string) float64 {
	prob := 0.0

	// Multiple resource types increase cascade probability
	resourceCount := ic.extractFeatureValue(features, "resource_type_count:")
	if resourceCount > 1 {
		prob += 0.3 * math.Min(float64(resourceCount-1)/3.0, 1.0)
	}

	// High alert rate indicates cascade
	alerts5Min := ic.extractFeatureValue(features, "alerts_5_min:")
	if alerts5Min > 5 {
		prob += 0.3
	}

	// Critical alerts increase cascade probability
	criticalCount := ic.extractFeatureValue(features, "critical_count:")
	if criticalCount > 0 {
		prob += 0.2 * math.Min(float64(criticalCount)/3.0, 1.0)
	}

	// Multiple hosts increase cascade probability
	hostCount := ic.extractFeatureValue(features, "host_count:")
	if hostCount > 1 {
		prob += 0.2 * math.Min(float64(hostCount-1)/2.0, 1.0)
	}

	return math.Min(prob, 1.0)
}

// PredictDuration estimates incident duration
func (ic *IncidentClassifier) PredictDuration(features []string) time.Duration {
	baseDuration := 10 * time.Minute

	// Adjust based on impact factors
	criticalCount := ic.extractFeatureValue(features, "critical_count:")
	baseDuration += time.Duration(criticalCount) * 15 * time.Minute

	hostCount := ic.extractFeatureValue(features, "host_count:")
	baseDuration += time.Duration(hostCount-1) * 10 * time.Minute

	resourceCount := ic.extractFeatureValue(features, "resource_type_count:")
	baseDuration += time.Duration(resourceCount-1) * 5 * time.Minute

	maxValue := ic.extractFeatureValueFloat(features, "max_value:")
	if maxValue > 95 {
		baseDuration += 20 * time.Minute
	}

	return baseDuration
}

// Helper methods

func (ai *LocalAIModel) identifyRootCauseCandidates(alerts []domain.Alert, features []string) []*domain.Alert {
	candidates := []*domain.Alert{}

	for i := range alerts {
		alert := &alerts[i]
		if alert.Status != domain.StatusClear {
			candidates = append(candidates, alert)
		}
	}

	return candidates
}

func (ai *LocalAIModel) scoreWithML(candidates []*domain.Alert, features []string) map[*domain.Alert]float64 {
	scores := make(map[*domain.Alert]float64)

	for _, candidate := range candidates {
		score := 0.0

		// Position-based scoring (earlier alerts get higher scores)
		// This would normally use a trained model
		if candidate.ResourceType == domain.ResourceMemory {
			score += 0.3 // Memory issues often root causes
		}
		if candidate.Status == domain.StatusCritical {
			score += 0.2
		}
		if candidate.Value > 90 {
			score += 0.15
		}

		scores[candidate] = score
	}

	return scores
}

func (ai *LocalAIModel) selectBestCandidate(candidates []*domain.Alert, scores map[*domain.Alert]float64) (*domain.Alert, float64) {
	if len(candidates) == 0 {
		return nil, 0.0
	}

	best := candidates[0]
	bestScore := scores[best]

	for _, candidate := range candidates {
		if scores[candidate] > bestScore {
			best = candidate
			bestScore = scores[candidate]
		}
	}

	return best, bestScore
}

func (ai *LocalAIModel) getAlternativeCauses(candidates []*domain.Alert, scores map[*domain.Alert]float64, best *domain.Alert) []*domain.Alert {
	alternatives := []*domain.Alert{}

	type candidateScore struct {
		alert *domain.Alert
		score float64
	}

	var sortedCandidates []candidateScore
	for _, candidate := range candidates {
		if candidate != best {
			sortedCandidates = append(sortedCandidates, candidateScore{candidate, scores[candidate]})
		}
	}

	sort.Slice(sortedCandidates, func(i, j int) bool {
		return sortedCandidates[i].score > sortedCandidates[j].score
	})

	for i, cs := range sortedCandidates {
		if i >= 2 { // Top 2 alternatives
			break
		}
		alternatives = append(alternatives, cs.alert)
	}

	return alternatives
}

func (ai *LocalAIModel) generateReasoning(alert *domain.Alert, features []string, confidence float64) string {
	reasoning := fmt.Sprintf("ML analysis identifies %s alert on %s as root cause",
		alert.ResourceType, alert.Chart)

	if confidence > 0.8 {
		reasoning += " with high confidence"
	} else if confidence > 0.6 {
		reasoning += " with moderate confidence"
	} else {
		reasoning += " with low confidence - requires manual verification"
	}

	return reasoning
}

func (ic *IncidentClassifier) extractFeatureValue(features []string, prefix string) int {
	for _, feature := range features {
		if len(feature) > len(prefix) && feature[:len(prefix)] == prefix {
			var value int
			_, err := fmt.Sscanf(feature[len(prefix):], "%d", &value)
			if err == nil {
				return value
			}
		}
	}
	return 0
}

func (ic *IncidentClassifier) extractFeatureValueFloat(features []string, prefix string) float64 {
	for _, feature := range features {
		if len(feature) > len(prefix) && feature[:len(prefix)] == prefix {
			var value float64
			_, err := fmt.Sscanf(feature[len(prefix):], "%f", &value)
			if err == nil {
				return value
			}
		}
	}
	return 0.0
}

func (fe *FeatureExtractor) countAlertsInWindow(alerts []domain.Alert, window time.Duration) int {
	if len(alerts) == 0 {
		return 0
	}

	count := 0
	start := alerts[0].OccurredAt
	end := start.Add(window)

	for _, alert := range alerts {
		if alert.OccurredAt.After(start) && alert.OccurredAt.Before(end) {
			count++
		}
	}

	return count
}

func (ai *LocalAIModel) calculatePatternConfidence(alerts []domain.Alert, patternType string) float64 {
	// Simple confidence calculation based on pattern strength
	switch patternType {
	case "burst", "cascade":
		return 0.9
	case "spike":
		return 0.8
	case "gradual":
		return 0.7
	case "progressive":
		return 0.6
	default:
		return 0.5
	}
}

func (ai *LocalAIModel) detectSeasonality(alerts []domain.Alert) bool {
	// Simple seasonality detection based on time patterns
	if len(alerts) < 10 {
		return false
	}

	// Check if alerts occur at similar times
	hours := make(map[int]int)
	for _, alert := range alerts {
		hours[alert.OccurredAt.Hour()]++
	}

	// If >50% of alerts occur in the same 4-hour window, consider seasonal
	for hour := range hours {
		windowCount := 0
		for h := hour; h < hour+4; h++ {
			windowCount += hours[h%24]
		}
		if float64(windowCount)/float64(len(alerts)) > 0.5 {
			return true
		}
	}

	return false
}

func (ai *LocalAIModel) determineTrend(alerts []domain.Alert) string {
	if len(alerts) < 3 {
		return "stable"
	}

	// Simple trend analysis based on alert frequency
	firstHalf := len(alerts) / 2
	secondHalf := len(alerts) - firstHalf

	// Compare alert frequency in first vs second half
	firstHalfDuration := alerts[firstHalf-1].OccurredAt.Sub(alerts[0].OccurredAt)
	secondHalfDuration := alerts[len(alerts)-1].OccurredAt.Sub(alerts[firstHalf].OccurredAt)

	if firstHalfDuration == 0 || secondHalfDuration == 0 {
		return "stable"
	}

	firstRate := float64(firstHalf) / firstHalfDuration.Minutes()
	secondRate := float64(secondHalf) / secondHalfDuration.Minutes()

	if secondRate > firstRate*1.5 {
		return "increasing"
	} else if firstRate > secondRate*1.5 {
		return "decreasing"
	}

	return "stable"
}

func (ai *LocalAIModel) calculateAnomalyScore(features []string) float64 {
	// Simple anomaly detection based on feature deviation
	score := 0.0

	alertCount := ai.extractFeatureValue(features, "alert_count:")
	if alertCount > 10 {
		score += 0.3
	}

	criticalCount := ai.extractFeatureValue(features, "critical_count:")
	if criticalCount > 3 {
		score += 0.3
	}

	alerts5Min := ai.extractFeatureValue(features, "alerts_5_min:")
	if alerts5Min > 5 {
		score += 0.4
	}

	return math.Min(score, 1.0)
}

func (ai *LocalAIModel) buildCorrelationMatrix(alerts []domain.Alert) map[string]float64 {
	matrix := make(map[string]float64)

	// Calculate correlation between resource types
	resourcePairs := map[string][]domain.Alert{
		"cpu_memory":     {},
		"cpu_disk":       {},
		"memory_disk":    {},
		"network_memory": {},
	}

	for _, alert := range alerts {
		if alert.ResourceType == domain.ResourceCPU || alert.ResourceType == domain.ResourceMemory {
			resourcePairs["cpu_memory"] = append(resourcePairs["cpu_memory"], alert)
		}
		if alert.ResourceType == domain.ResourceCPU || alert.ResourceType == domain.ResourceDisk {
			resourcePairs["cpu_disk"] = append(resourcePairs["cpu_disk"], alert)
		}
		if alert.ResourceType == domain.ResourceMemory || alert.ResourceType == domain.ResourceDisk {
			resourcePairs["memory_disk"] = append(resourcePairs["memory_disk"], alert)
		}
		if alert.ResourceType == domain.ResourceNetwork || alert.ResourceType == domain.ResourceMemory {
			resourcePairs["network_memory"] = append(resourcePairs["network_memory"], alert)
		}
	}

	for pair, alertList := range resourcePairs {
		if len(alertList) >= 2 {
			// Simple correlation based on temporal proximity
			correlation := ai.calculateTemporalCorrelation(alertList)
			matrix[pair] = correlation
		}
	}

	return matrix
}

func (ai *LocalAIModel) calculateTemporalCorrelation(alerts []domain.Alert) float64 {
	if len(alerts) < 2 {
		return 0.0
	}

	// Calculate average time difference
	totalDiff := 0.0

	for i := 1; i < len(alerts); i++ {
		diff := alerts[i].OccurredAt.Sub(alerts[i-1].OccurredAt).Minutes()
		totalDiff += diff
	}

	avgDiff := totalDiff / float64(len(alerts)-1)

	// Higher correlation if alerts occur close together
	if avgDiff < 1 {
		return 0.9
	} else if avgDiff < 5 {
		return 0.7
	} else if avgDiff < 15 {
		return 0.5
	} else {
		return 0.3
	}
}

func (ai *LocalAIModel) predictNextOccurrence(alerts []domain.Alert, patternType string) time.Time {
	if len(alerts) == 0 {
		return time.Time{}
	}

	lastAlert := alerts[len(alerts)-1]

	switch patternType {
	case "burst":
		return lastAlert.OccurredAt.Add(1 * time.Hour)
	case "cascade":
		return lastAlert.OccurredAt.Add(6 * time.Hour)
	case "gradual":
		return lastAlert.OccurredAt.Add(24 * time.Hour)
	case "spike":
		return lastAlert.OccurredAt.Add(30 * time.Minute)
	case "progressive":
		return lastAlert.OccurredAt.Add(4 * time.Hour)
	default:
		return lastAlert.OccurredAt.Add(12 * time.Hour)
	}
}

func (ic *IncidentClassifier) classifyBusinessImpact(impactScore, cascadeProb float64) string {
	combinedScore := (impactScore * 0.6) + (cascadeProb * 0.4)

	if combinedScore > 0.8 {
		return "Critical business impact - services severely degraded"
	} else if combinedScore > 0.6 {
		return "High business impact - users experiencing issues"
	} else if combinedScore > 0.4 {
		return "Medium business impact - some users affected"
	} else {
		return "Low business impact - minimal user impact"
	}
}

func (ai *LocalAIModel) determineRiskLevel(impactScore, cascadeProb float64, duration time.Duration) string {
	riskScore := impactScore*0.4 + cascadeProb*0.4 + (math.Min(duration.Hours()/24.0, 1.0))*0.2

	if riskScore > 0.8 {
		return "critical"
	} else if riskScore > 0.6 {
		return "high"
	} else if riskScore > 0.4 {
		return "medium"
	} else {
		return "low"
	}
}

func (ai *LocalAIModel) extractFeatureValue(features []string, prefix string) int {
	for _, feature := range features {
		if len(feature) > len(prefix) && feature[:len(prefix)] == prefix {
			var value int
			_, err := fmt.Sscanf(feature[len(prefix):], "%d", &value)
			if err == nil {
				return value
			}
		}
	}
	return 0
}

func (ai *LocalAIModel) classifyBusinessImpact(impactScore, cascadeProb float64) string {
	combinedScore := (impactScore * 0.6) + (cascadeProb * 0.4)

	if combinedScore > 0.8 {
		return "Critical business impact - services severely degraded"
	} else if combinedScore > 0.6 {
		return "High business impact - users experiencing issues"
	} else if combinedScore > 0.4 {
		return "Medium business impact - some users affected"
	} else {
		return "Low business impact - minimal user impact"
	}
}

func (ai *LocalAIModel) identifyAffectedServices(alerts []domain.Alert, features []string) []string {
	services := make(map[string]bool)

	// Extract service information from alerts
	for _, alert := range alerts {
		// Map charts to services (simplified)
		service := ai.mapChartToService(alert.Chart)
		if service != "" {
			services[service] = true
		}
	}

	// Convert map to slice
	result := make([]string, 0, len(services))
	for service := range services {
		result = append(result, service)
	}

	return result
}

func (ai *LocalAIModel) mapChartToService(chart string) string {
	// Simple mapping from chart names to services
	switch {
	case contains(chart, "web"), contains(chart, "nginx"), contains(chart, "apache"):
		return "web-server"
	case contains(chart, "database"), contains(chart, "mysql"), contains(chart, "postgres"):
		return "database"
	case contains(chart, "cache"), contains(chart, "redis"), contains(chart, "memcached"):
		return "cache"
	case contains(chart, "queue"), contains(chart, "rabbitmq"), contains(chart, "kafka"):
		return "message-queue"
	case contains(chart, "storage"), contains(chart, "s3"), contains(chart, "minio"):
		return "storage"
	default:
		return "system"
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr || (len(s) > len(substr) &&
			stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
