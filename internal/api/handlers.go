package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"incident-teller/internal/ai"
	"incident-teller/internal/domain"
	"incident-teller/internal/observability"
)

// Handler provides HTTP handlers for the IncidentTeller API
type Handler struct {
	repo          Repository
	aiModel       ai.AIModel
	logger        observability.Logger
	healthChecker observability.HealthChecker
}

// Repository interface for data access
type Repository interface {
	SaveAlert(ctx context.Context, alert domain.Alert) error
	GetIncidents(ctx context.Context) ([]domain.Incident, error)
	GetLastProcessedID(ctx context.Context) (uint64, error)
	SetLastProcessedID(ctx context.Context, id uint64) error
}

// NewHandler creates a new API handler
func NewHandler(repo Repository, aiModel ai.AIModel, logger observability.Logger, healthChecker observability.HealthChecker) *Handler {
	return &Handler{
		repo:          repo,
		aiModel:       aiModel,
		logger:        logger,
		healthChecker: healthChecker,
	}
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Code    int    `json:"code"`
}

// IncidentSummaryResponse represents the summary statistics
type IncidentSummaryResponse struct {
	ActiveIncidents   int     `json:"active_incidents"`
	ResolvedIncidents int     `json:"resolved_incidents"`
	AverageConfidence float64 `json:"average_confidence"`
	RiskLevel         string  `json:"risk_level"`
	LastIncidentTime  *string `json:"last_incident_time,omitempty"`
}

// IncidentDetailResponse represents a single incident with AI analysis
type IncidentDetailResponse struct {
	ID            string                  `json:"id"`
	Title         string                  `json:"title"`
	Status        string                  `json:"status"`
	StartedAt     time.Time               `json:"started_at"`
	ResolvedAt    *time.Time              `json:"resolved_at,omitempty"`
	Duration      string                  `json:"duration"`
	RootCause     *RootCauseResponse      `json:"root_cause,omitempty"`
	BlastRadius   *BlastRadiusResponse    `json:"blast_radius,omitempty"`
	RiskLevel     string                  `json:"risk_level"`
	TotalEvents   int                     `json:"total_events"`
	EventTimeline []TimelineEventResponse `json:"event_timeline"`
}

// RootCauseResponse represents AI root cause analysis
type RootCauseResponse struct {
	AlertID           string                     `json:"alert_id"`
	ResourceType      string                     `json:"resource_type"`
	Chart             string                     `json:"chart"`
	Host              string                     `json:"host"`
	Confidence        float64                    `json:"confidence"`
	PatternType       string                     `json:"pattern_type"`
	Reasoning         string                     `json:"reasoning"`
	AlternativeCauses []AlternativeCauseResponse `json:"alternative_causes"`
}

// AlternativeCauseResponse represents alternative root causes
type AlternativeCauseResponse struct {
	AlertID      string  `json:"alert_id"`
	ResourceType string  `json:"resource_type"`
	Chart        string  `json:"chart"`
	Host         string  `json:"host"`
	Confidence   float64 `json:"confidence"`
}

// BlastRadiusResponse represents blast radius analysis
type BlastRadiusResponse struct {
	ImpactScore        float64  `json:"impact_score"`
	AffectedServices   []string `json:"affected_services"`
	CascadeProbability float64  `json:"cascade_probability"`
	DurationPredicted  string   `json:"duration_predicted"`
	BusinessImpact     string   `json:"business_impact"`
	RiskLevel          string   `json:"risk_level"`
}

// TimelineEventResponse represents a timeline event
type TimelineEventResponse struct {
	Timestamp          time.Time `json:"timestamp"`
	Type               string    `json:"type"`
	Message            string    `json:"message"`
	Severity           string    `json:"severity"`
	DurationSinceStart *string   `json:"duration_since_start,omitempty"`
	ResourceType       string    `json:"resource_type"`
}

// IncidentListResponse represents a list of incidents
type IncidentListResponse struct {
	Incidents []IncidentListItemResponse `json:"incidents"`
	Total     int                        `json:"total"`
	Page      int                        `json:"page"`
	PageSize  int                        `json:"page_size"`
}

// IncidentListItemResponse represents a single incident in a list
type IncidentListItemResponse struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Status      string     `json:"status"`
	StartedAt   time.Time  `json:"started_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
	Duration    string     `json:"duration"`
	RootCause   string     `json:"root_cause"`
	TotalEvents int        `json:"total_events"`
	RiskLevel   string     `json:"risk_level"`
}

// HealthResponse represents health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Version   string            `json:"version"`
	Timestamp time.Time         `json:"timestamp"`
	Checks    map[string]string `json:"checks,omitempty"`
}

// SetupRoutes configures the API routes
func (h *Handler) SetupRoutes() *http.ServeMux {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/incidents/summary", h.handleIncidentsSummary)
	mux.HandleFunc("/api/incidents", h.handleIncidents)
	mux.HandleFunc("/api/incidents/", h.handleIncidentDetail)
	mux.HandleFunc("/api/health", h.handleHealth)

	return mux
}

// handleIncidentsSummary returns incident summary statistics
func (h *Handler) handleIncidentsSummary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	ctx := r.Context()

	incidents, err := h.repo.GetIncidents(ctx)
	if err != nil {
		h.logger.Error("Failed to get incidents for summary", observability.Error(err))
		h.writeError(w, http.StatusInternalServerError, "Failed to get incidents")
		return
	}

	activeIncidents := 0
	resolvedIncidents := 0
	var totalConfidence float64
	confidenceCount := 0
	var lastIncidentTime *time.Time
	riskLevels := make(map[string]int)

	for _, incident := range incidents {
		if incident.ResolvedAt == nil {
			activeIncidents++
		} else {
			resolvedIncidents++
		}

		// Calculate risk level based on incident characteristics
		riskLevel := h.calculateRiskLevel(incident)
		riskLevels[riskLevel]++

		// Get AI analysis for confidence scores
		if h.aiModel != nil && len(incident.Events) > 0 {
			rootCause, err := h.aiModel.PredictRootCause(ctx, incident.Events)
			if err == nil {
				totalConfidence += rootCause.Confidence
				confidenceCount++
			}
		}

		// Track last incident time
		if lastIncidentTime == nil || incident.StartedAt.After(*lastIncidentTime) {
			lastIncidentTime = &incident.StartedAt
		}
	}

	avgConfidence := 0.0
	if confidenceCount > 0 {
		avgConfidence = totalConfidence / float64(confidenceCount)
	}

	// Determine overall risk level
	overallRiskLevel := "low"
	if riskLevels["critical"] > 0 || riskLevels["high"] > 0 {
		overallRiskLevel = "high"
	} else if riskLevels["medium"] > 0 {
		overallRiskLevel = "medium"
	}

	response := IncidentSummaryResponse{
		ActiveIncidents:   activeIncidents,
		ResolvedIncidents: resolvedIncidents,
		AverageConfidence: avgConfidence,
		RiskLevel:         overallRiskLevel,
	}

	if lastIncidentTime != nil {
		formatted := lastIncidentTime.Format(time.RFC3339)
		response.LastIncidentTime = &formatted
	}

	h.writeJSON(w, http.StatusOK, response)
}

// handleIncidents returns a list of incidents
func (h *Handler) handleIncidents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	ctx := r.Context()

	incidents, err := h.repo.GetIncidents(ctx)
	if err != nil {
		h.logger.Error("Failed to get incidents", observability.Error(err))
		h.writeError(w, http.StatusInternalServerError, "Failed to get incidents")
		return
	}

	// Parse query parameters
	page := 1
	pageSize := 20

	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	// Convert to response format
	var incidentItems []IncidentListItemResponse
	for _, incident := range incidents {
		rootCause := h.identifyPrimaryRootCause(incident)
		riskLevel := h.calculateRiskLevel(incident)
		duration := h.calculateDuration(incident)

		item := IncidentListItemResponse{
			ID:          incident.ID,
			Title:       incident.Title,
			Status:      string(incident.Status),
			StartedAt:   incident.StartedAt,
			ResolvedAt:  incident.ResolvedAt,
			Duration:    duration,
			RootCause:   rootCause,
			TotalEvents: len(incident.Events),
			RiskLevel:   riskLevel,
		}
		incidentItems = append(incidentItems, item)
	}

	// Pagination
	total := len(incidentItems)
	start := (page - 1) * pageSize
	end := start + pageSize
	if start >= total {
		incidentItems = []IncidentListItemResponse{}
	} else if end > total {
		incidentItems = incidentItems[start:]
	} else {
		incidentItems = incidentItems[start:end]
	}

	response := IncidentListResponse{
		Incidents: incidentItems,
		Total:     total,
		Page:      page,
		PageSize:  pageSize,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// handleIncidentDetail returns detailed information about a specific incident
func (h *Handler) handleIncidentDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract incident ID from URL
	id := extractIncidentID(r.URL.Path)
	if id == "" {
		h.writeError(w, http.StatusBadRequest, "Invalid incident ID")
		return
	}

	ctx := r.Context()

	incidents, err := h.repo.GetIncidents(ctx)
	if err != nil {
		h.logger.Error("Failed to get incidents", observability.Error(err))
		h.writeError(w, http.StatusInternalServerError, "Failed to get incidents")
		return
	}

	// Find the specific incident
	var incident *domain.Incident
	for i, inc := range incidents {
		if inc.ID == id {
			incident = &incidents[i]
			break
		}
	}

	if incident == nil {
		h.writeError(w, http.StatusNotFound, "Incident not found")
		return
	}

	// Perform AI analysis
	var rootCauseResponse *RootCauseResponse
	var blastRadiusResponse *BlastRadiusResponse

	if h.aiModel != nil && len(incident.Events) > 0 {
		if rootCause, err := h.aiModel.PredictRootCause(ctx, incident.Events); err == nil {
			rootCauseResponse = h.convertRootCauseToResponse(rootCause)
		}

		if blastRadius, err := h.aiModel.PredictBlastRadius(ctx, incident.Events); err == nil {
			blastRadiusResponse = h.convertBlastRadiusToResponse(blastRadius)
		}
	}

	response := IncidentDetailResponse{
		ID:            incident.ID,
		Title:         incident.Title,
		Status:        string(incident.Status),
		StartedAt:     incident.StartedAt,
		ResolvedAt:    incident.ResolvedAt,
		Duration:      h.calculateDuration(*incident),
		RootCause:     rootCauseResponse,
		BlastRadius:   blastRadiusResponse,
		RiskLevel:     h.calculateRiskLevel(*incident),
		TotalEvents:   len(incident.Events),
		EventTimeline: h.convertTimelineToResponse(incident),
	}

	h.writeJSON(w, http.StatusOK, response)
}

// handleHealth returns system health information
func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	health := h.healthChecker.CheckHealth(r.Context())

	response := HealthResponse{
		Status:    health.Status,
		Version:   health.Version,
		Timestamp: health.Timestamp,
		Checks:    make(map[string]string),
	}

	// Add individual check statuses
	for name, check := range health.Checks {
		response.Checks[name] = check.Status
	}

	statusCode := http.StatusOK
	if health.Status != "healthy" {
		statusCode = http.StatusServiceUnavailable
	}

	h.writeJSON(w, statusCode, response)
}

// Helper methods

func (h *Handler) writeError(w http.ResponseWriter, code int, message string) {
	h.writeJSON(w, code, ErrorResponse{
		Error:   http.StatusText(code),
		Message: message,
		Code:    code,
	})
}

func (h *Handler) writeJSON(w http.ResponseWriter, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", observability.Error(err))
	}
}

func extractIncidentID(path string) string {
	// Extract ID from /api/incidents/{id}
	prefix := "/api/incidents/"
	if len(path) <= len(prefix) {
		return ""
	}

	id := path[len(prefix):]

	// Remove any query parameters
	if idx := strings.Index(id, "?"); idx != -1 {
		id = id[:idx]
	}

	return id
}

func (h *Handler) identifyPrimaryRootCause(incident domain.Incident) string {
	if len(incident.Events) == 0 {
		return "Unknown"
	}

	// Find the first critical or warning alert
	for _, event := range incident.Events {
		if event.Status == domain.StatusCritical || event.Status == domain.StatusWarning {
			return string(event.ResourceType)
		}
	}

	return string(incident.Events[0].ResourceType)
}

func (h *Handler) calculateRiskLevel(incident domain.Incident) string {
	if len(incident.Events) == 0 {
		return "low"
	}

	criticalCount := 0
	hostCount := make(map[string]bool)
	resourceTypes := make(map[domain.ResourceType]bool)

	for _, event := range incident.Events {
		if event.Status == domain.StatusCritical {
			criticalCount++
		}
		hostCount[event.Host] = true
		resourceTypes[event.ResourceType] = true
	}

	// Calculate risk based on severity, scope, and duration
	if criticalCount >= 3 || len(hostCount) >= 3 || len(resourceTypes) >= 3 {
		return "critical"
	} else if criticalCount >= 2 || len(hostCount) >= 2 || len(resourceTypes) >= 2 {
		return "high"
	} else if criticalCount >= 1 || len(hostCount) > 1 {
		return "medium"
	}

	return "low"
}

func (h *Handler) calculateDuration(incident domain.Incident) string {
	if incident.ResolvedAt == nil {
		return time.Since(incident.StartedAt).String() + " (ongoing)"
	}
	return incident.ResolvedAt.Sub(incident.StartedAt).String()
}

func (h *Handler) convertRootCauseToResponse(rootCause ai.RootCausePrediction) *RootCauseResponse {
	if rootCause.PrimaryCause == nil {
		return &RootCauseResponse{
			AlertID:           "",
			ResourceType:      "unknown",
			Chart:             "",
			Host:              "",
			Confidence:        0.0,
			PatternType:       rootCause.PatternType,
			Reasoning:         rootCause.Reasoning,
			AlternativeCauses: []AlternativeCauseResponse{},
		}
	}

	var alternatives []AlternativeCauseResponse
	for _, alt := range rootCause.AlternativeCauses {
		alternatives = append(alternatives, AlternativeCauseResponse{
			AlertID:      alt.ID,
			ResourceType: string(alt.ResourceType),
			Chart:        alt.Chart,
			Host:         alt.Host,
			Confidence:   0.0, // Would need scoring logic
		})
	}

	return &RootCauseResponse{
		AlertID:           rootCause.PrimaryCause.ID,
		ResourceType:      string(rootCause.PrimaryCause.ResourceType),
		Chart:             rootCause.PrimaryCause.Chart,
		Host:              rootCause.PrimaryCause.Host,
		Confidence:        rootCause.Confidence,
		PatternType:       rootCause.PatternType,
		Reasoning:         rootCause.Reasoning,
		AlternativeCauses: alternatives,
	}
}

func (h *Handler) convertBlastRadiusToResponse(blastRadius ai.BlastRadiusPrediction) *BlastRadiusResponse {
	return &BlastRadiusResponse{
		ImpactScore:        blastRadius.ImpactScore,
		AffectedServices:   blastRadius.AffectedServices,
		CascadeProbability: blastRadius.CascadeProbability,
		DurationPredicted:  blastRadius.DurationPredicted.String(),
		BusinessImpact:     blastRadius.BusinessImpact,
		RiskLevel:          blastRadius.RiskLevel,
	}
}

func (h *Handler) convertTimelineToResponse(incident *domain.Incident) []TimelineEventResponse {
	var timeline []TimelineEventResponse

	for i, event := range incident.Events {
		var durationSinceStart *string
		if i > 0 {
			duration := event.OccurredAt.Sub(incident.Events[0].OccurredAt).String()
			durationSinceStart = &duration
		}

		eventType := "TRIGGERED"
		if event.Status == domain.StatusClear {
			eventType = "RESOLVED"
		} else if i > 0 && event.Status != incident.Events[i-1].Status {
			eventType = "UPDATE"
		}

		severity := "info"
		if event.Status == domain.StatusCritical {
			severity = "critical"
		} else if event.Status == domain.StatusWarning {
			severity = "warning"
		}

		message := h.generateEventMessage(event)

		timeline = append(timeline, TimelineEventResponse{
			Timestamp:          event.OccurredAt,
			Type:               eventType,
			Message:            message,
			Severity:           severity,
			DurationSinceStart: durationSinceStart,
			ResourceType:       string(event.ResourceType),
		})
	}

	return timeline
}

func (h *Handler) generateEventMessage(event domain.Alert) string {
	switch event.Status {
	case domain.StatusCritical:
		return fmt.Sprintf("Critical alert triggered for %s on %s (value: %.2f)",
			event.ResourceType, event.Host, event.Value)
	case domain.StatusWarning:
		return fmt.Sprintf("Warning alert for %s on %s (value: %.2f)",
			event.ResourceType, event.Host, event.Value)
	case domain.StatusClear:
		return fmt.Sprintf("%s alert resolved on %s",
			event.ResourceType, event.Host)
	default:
		return fmt.Sprintf("Alert status changed to %s for %s on %s",
			event.Status, event.ResourceType, event.Host)
	}
}
