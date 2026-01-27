package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"incident-teller/internal/ai"
	"incident-teller/internal/domain"
	"incident-teller/internal/observability"
	"incident-teller/internal/services"
)

// Handler provides HTTP handlers for the IncidentTeller API
type Handler struct {
	repo          Repository
	aiModel       ai.AIModel
	logger        observability.Logger
	healthChecker observability.HealthChecker
	metrics       observability.Metrics
}

// Repository interface for data access
type Repository interface {
	SaveAlert(ctx context.Context, alert domain.Alert) error
	GetIncidents(ctx context.Context) ([]domain.Incident, error)
	GetLastProcessedID(ctx context.Context) (uint64, error)
	SetLastProcessedID(ctx context.Context, id uint64) error
	SaveIncident(ctx context.Context, incident domain.Incident) error
	GetAlerts(ctx context.Context) ([]domain.Alert, error)
	Stats(ctx context.Context) (map[string]interface{}, error)
	PingContext(ctx context.Context) error
}

// NewHandler creates a new API handler
func NewHandler(repo Repository, aiModel ai.AIModel, logger observability.Logger, healthChecker observability.HealthChecker, metrics observability.Metrics) *Handler {
	return &Handler{
		repo:          repo,
		aiModel:       aiModel,
		logger:        logger,
		healthChecker: healthChecker,
		metrics:       metrics,
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

// AIAnalysisResponse represents AI-generated insights
type AIAnalysisResponse struct {
	Summary         string                 `json:"summary"`
	RootCauseText   string                 `json:"root_cause_text"`
	ImpactAssessment string                `json:"impact_assessment"`
	Recommendations RecommendationsResponse `json:"recommendations"`
	GeneratedAt     time.Time              `json:"generated_at"`
	AlertCount      int                    `json:"alert_count"`
	TimeSpan        string                 `json:"time_span"`
}

// RecommendationsResponse contains actionable recommendations
type RecommendationsResponse struct {
	Immediate []string `json:"immediate"`
	ShortTerm []string `json:"short_term"`
	LongTerm  []string `json:"long_term"`
}

// AlertGroupResponse represents a group of related alerts
type AlertGroupResponse struct {
	ID              string          `json:"id"`
	AlertCount      int             `json:"alert_count"`
	PrimaryHost     string          `json:"primary_host"`
	AffectedHosts   []string        `json:"affected_hosts"`
	ResourceTypes   []string        `json:"resource_types"`
	StartTime       time.Time       `json:"start_time"`
	EndTime         time.Time       `json:"end_time"`
	Duration        string          `json:"duration"`
	IsCascading     bool            `json:"is_cascading"`
	GroupType       string          `json:"group_type"`
	Alerts          []domain.Alert  `json:"alerts"`
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

// TimelineResponse represents a timeline response
type TimelineResponse struct {
	IncidentID string                  `json:"incident_id"`
	Events     []TimelineEventResponse `json:"events"`
	Total      int                     `json:"total"`
	Duration   string                  `json:"duration"`
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

// SetupRoutes configures the API routes and applies middleware
func (h *Handler) SetupRoutes() http.Handler {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("/api/incidents/summary", h.handleIncidentsSummary)
	mux.HandleFunc("/api/incidents", h.handleIncidents)
	mux.HandleFunc("/api/incidents/", h.handleIncidentDetail)
	mux.HandleFunc("/api/timeline/", h.handleIncidentTimeline)
	mux.HandleFunc("/api/timeline-enhanced/", h.handleIncidentTimelineEnhanced)
	mux.HandleFunc("/api/health", h.handleHealth)
	mux.HandleFunc("/api/logs", h.handleLogs)
	mux.HandleFunc("/api/metrics/export", h.handleMetricsExport)
	mux.HandleFunc("/api/diagnostics", h.handleDiagnostics)
	mux.HandleFunc("/api/events", h.handleSSE)
	mux.HandleFunc("/api/test/create-incident", h.handleCreateTestIncident)
	
	// AI-powered analysis endpoints
	mux.HandleFunc("/api/analyze", h.handleAIAnalysis)
	mux.HandleFunc("/api/alert-groups", h.handleAlertGroups)

	return h.withCORS(mux)
}

// withCORS is a middleware that handles Cross-Origin Resource Sharing
func (h *Handler) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// handleLogs returns the recent buffered logs
func (h *Handler) handleLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	logs := h.logger.GetLogs()
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"logs":  logs,
		"count": len(logs),
	})
}

// handleMetricsExport returns metrics in CSV format
func (h *Handler) handleMetricsExport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment;filename=metrics.csv")
	w.WriteHeader(http.StatusOK)

	fmt.Fprintf(w, "Metric,Value,Type\n")

	if m, ok := h.metrics.(*observability.StandardMetrics); ok {
		for k, v := range m.GetCounters() {
			fmt.Fprintf(w, "%s,%.2f,counter\n", k, v)
		}
		for k, v := range m.GetGauges() {
			fmt.Fprintf(w, "%s,%.2f,gauge\n", k, v)
		}
	}
}

// handleDiagnostics returns system diagnostic results
func (h *Handler) handleDiagnostics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	ctx := r.Context()
	health := h.healthChecker.CheckHealth(ctx)
	repoStats, _ := h.repo.Stats(ctx)

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	diagnostics := []map[string]interface{}{
		{
			"check":   "database_connectivity",
			"status":  health.Checks["database"].Status,
			"details": fmt.Sprintf("Records: %v alerts, %v incidents", repoStats["total_alerts"], repoStats["total_incidents"]),
		},
		{
			"check":   "netdata_api_connectivity",
			"status":  health.Checks["netdata"].Status,
			"details": health.Checks["netdata"].Message,
		},
		{
			"check":   "process_memory",
			"status":  "pass",
			"details": fmt.Sprintf("Alloc: %v MB, Sys: %v MB", mem.Alloc/1024/1024, mem.Sys/1024/1024),
		},
		{
			"check":   "incident_correlation_engine",
			"status":  "pass",
			"details": fmt.Sprintf("Last processed ID: %v", repoStats["last_processed_id"]),
		},
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":      health.Status,
		"diagnostics": diagnostics,
		"timestamp":   time.Now(),
	})
}

// handleSSE provides Server-Sent Events for real-time updates
func (h *Handler) handleSSE(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	flusher, ok := w.(http.Flusher)
	if !ok {
		h.logger.Error("Streaming unsupported")
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	// Send initial data
	h.sendSSEUpdate(w, flusher, ctx)

	for {
		select {
		case <-ctx.Done():
			h.logger.Info("SSE client disconnected")
			return
		case <-ticker.C:
			h.sendSSEUpdate(w, flusher, ctx)
		}
	}
}

// sendSSEUpdate sends current incidents data via SSE
func (h *Handler) sendSSEUpdate(w http.ResponseWriter, flusher http.Flusher, ctx context.Context) {
	incidents, err := h.repo.GetIncidents(ctx)
	if err != nil {
		h.logger.Error("Failed to get incidents for SSE", observability.Error(err))
		return
	}

	if len(incidents) == 0 {
		return
	}

	// Send the latest incident
	latest := incidents[len(incidents)-1]
	data, err := json.Marshal(latest)
	if err != nil {
		h.logger.Error("Failed to marshal incident for SSE", observability.Error(err))
		return
	}

	fmt.Fprintf(w, "data: %s\n\n", data)
	flusher.Flush()
}

// handleCreateTestIncident creates a test incident for development
func (h *Handler) handleCreateTestIncident(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	ctx := r.Context()

	// Create a test critical alert
	alert := domain.Alert{
		ID:           fmt.Sprintf("test-%d", time.Now().Unix()),
		ExternalID:   uint64(time.Now().Unix()),
		Host:         "localhost",
		Chart:        "system.cpu",
		Family:       "cpu",
		Name:         "high_cpu_usage",
		Status:       domain.StatusCritical,
		OldStatus:    domain.StatusClear,
		Value:        95.0,
		OccurredAt:   time.Now(),
		Description:  "Test critical CPU alert",
		ResourceType: domain.ResourceCPU,
		Labels: map[string]string{
			"source": "test",
		},
	}

	// Save the alert
	if err := h.repo.SaveAlert(ctx, alert); err != nil {
		h.logger.Error("Failed to save test alert", observability.Error(err))
		h.writeError(w, http.StatusInternalServerError, "Failed to save alert")
		return
	}

	// Create incident from this alert
	builder := services.NewIncidentBuilder(15 * time.Minute)

	// Get all alerts and build incidents
	alerts, err := h.repo.GetAlerts(ctx)
	if err != nil {
		h.logger.Error("Failed to get alerts", observability.Error(err))
		h.writeError(w, http.StatusInternalServerError, "Failed to get alerts")
		return
	}

	incidents := builder.Build(alerts)

	// Save the new incidents
	for _, incident := range incidents {
		if err := h.repo.SaveIncident(ctx, incident); err != nil {
			h.logger.Error("Failed to save incident", observability.Error(err))
			h.writeError(w, http.StatusInternalServerError, "Failed to save incident")
			return
		}
	}

	if len(incidents) > 0 {
		h.logger.Info("Test incident created",
			observability.String("incident_id", incidents[0].ID),
			observability.Int("alert_count", len(incidents[0].Events)))
	}

	h.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"incident_count": len(incidents),
		"alert_id":       alert.ID,
		"message":        "Test incident created successfully",
	})
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

	// Always return 200 OK so the frontend can see the health status
	// This allows the frontend to display health information even if some checks fail
	h.writeJSON(w, http.StatusOK, response)
}

// handleIncidentTimeline returns timeline events for a specific incident
func (h *Handler) handleIncidentTimeline(w http.ResponseWriter, r *http.Request) {
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

	// Convert timeline to response format
	timelineEvents := h.convertTimelineToResponse(incident)

	// Calculate incident duration
	duration := h.calculateDuration(*incident)

	response := TimelineResponse{
		IncidentID: incident.ID,
		Events:     timelineEvents,
		Total:      len(timelineEvents),
		Duration:   duration,
	}

	h.writeJSON(w, http.StatusOK, response)
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

// handleAIAnalysis generates AI-powered insights for all current alerts
func (h *Handler) handleAIAnalysis(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get all alerts
	alerts, err := h.repo.GetAlerts(ctx)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get alerts: %v", err))
		return
	}

	if len(alerts) == 0 {
		h.writeError(w, http.StatusBadRequest, "No alerts available for analysis")
		return
	}

	// Get AI analysis
	analysisData, err := h.getAIAnalysis(ctx, alerts)
	if err != nil {
		h.logger.Error("Failed to generate AI analysis", observability.Field{Key: "error", Value: err})
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to generate analysis: %v", err))
		return
	}

	// Convert interface{} to map
	analysisMap, ok := analysisData.(map[string]interface{})
	if !ok {
		h.logger.Error("Invalid analysis response format", observability.Field{Key: "type", Value: fmt.Sprintf("%T", analysisData)})
		h.writeError(w, http.StatusInternalServerError, "Invalid analysis format")
		return
	}

	// Extract recommendation data
	var recommendations RecommendationsResponse
	if rec, ok := analysisMap["recommendations"].(map[string]interface{}); ok {
		if imm, ok := rec["immediate"].([]string); ok {
			recommendations.Immediate = imm
		}
		if st, ok := rec["short_term"].([]string); ok {
			recommendations.ShortTerm = st
		}
		if lt, ok := rec["long_term"].([]string); ok {
			recommendations.LongTerm = lt
		}
	}

	// Build response
	response := AIAnalysisResponse{
		Summary:          fmt.Sprintf("%v", analysisMap["summary"]),
		RootCauseText:    fmt.Sprintf("%v", analysisMap["root_cause"]),
		ImpactAssessment: fmt.Sprintf("%v", analysisMap["impact"]),
		Recommendations:  recommendations,
		GeneratedAt:      time.Now(),
		AlertCount:       len(alerts),
		TimeSpan:         "incident analysis",
	}

	h.writeJSON(w, http.StatusOK, response)
}

// handleAlertGroups returns alerts grouped by host and cascade relationships
func (h *Handler) handleAlertGroups(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get all alerts
	alerts, err := h.repo.GetAlerts(ctx)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get alerts: %v", err))
		return
	}

	// Group alerts
	grouper := services.NewAlertGrouper(15 * time.Minute)
	groups := grouper.GroupAlerts(alerts)

	// Convert to response format
	groupResponses := make([]AlertGroupResponse, len(groups))
	for i, group := range groups {
		resourceTypes := make([]string, len(group.ResourceTypes))
		for j, rt := range group.ResourceTypes {
			resourceTypes[j] = string(rt)
		}

		groupResponses[i] = AlertGroupResponse{
			ID:            group.ID,
			AlertCount:    len(group.Alerts),
			PrimaryHost:   group.PrimaryHost,
			AffectedHosts: group.AffectedHosts,
			ResourceTypes: resourceTypes,
			StartTime:     group.StartTime,
			EndTime:       group.EndTime,
			Duration:      group.EndTime.Sub(group.StartTime).String(),
			IsCascading:   group.IsCascading,
			GroupType:     group.GroupType,
			Alerts:        group.Alerts,
		}
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"groups": groupResponses,
		"total":  len(groupResponses),
	})
}

// handleIncidentTimelineEnhanced returns an enhanced timeline with cascade detection
func (h *Handler) handleIncidentTimelineEnhanced(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	incidentID := r.PathValue("id")
	if incidentID == "" {
		h.writeError(w, http.StatusBadRequest, "Missing incident ID")
		return
	}

	// Get incident
	incidents, err := h.repo.GetIncidents(ctx)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to get incidents: %v", err))
		return
	}

	var incident *domain.Incident
	for _, inc := range incidents {
		if inc.ID == incidentID {
			incident = &inc
			break
		}
	}

	if incident == nil {
		h.writeError(w, http.StatusNotFound, "Incident not found")
		return
	}

	// Group alerts and build enhanced timeline
	grouper := services.NewAlertGrouper(15 * time.Minute)
	groups := grouper.GroupAlerts(incident.Events)

	timelineBuilder := services.NewEnhancedTimelineBuilder(grouper)
	timeline := timelineBuilder.BuildTimeline(incident.Events, groups)

	// Convert to response format
	eventResponses := make([]map[string]interface{}, len(timeline.Events))
	for i, event := range timeline.Events {
		relativeTime := event.TimeFromIncidentStart.String()
		eventResponses[i] = map[string]interface{}{
			"timestamp":             event.Timestamp,
			"type":                  event.Type,
			"severity":              event.Severity,
			"message":               event.Message,
			"duration_since_start":  relativeTime,
			"is_cascade_point":      event.IsCascadePoint,
			"is_root_cause":         event.IsRootCause,
			"resources_affected":    event.ResourcesAffected,
		}
	}

	response := map[string]interface{}{
		"incident_id":              incident.ID,
		"events":                   eventResponses,
		"total_events":             len(timeline.Events),
		"duration":                 timeline.Duration.String(),
		"start_time":               timeline.StartTime,
		"end_time":                 timeline.EndTime,
		"critical_points":          timeline.CriticalPoints,
		"root_cause_event_index":   timeline.RootCauseEventIndex,
		"resolution_event_index":   timeline.ResolutionEventIndex,
	}

	h.writeJSON(w, http.StatusOK, response)
}

// getAIAnalysis gets AI-powered analysis (integrating with OpenAI)
func (h *Handler) getAIAnalysis(ctx context.Context, alerts []domain.Alert) (interface{}, error) {
	// Try to use OpenAI if available
	openaiClient, err := h.getOpenAIClient()
	if err == nil && openaiClient != nil {
		return h.getOpenAIAnalysis(ctx, openaiClient, alerts)
	}

	// Fall back to local analysis
	return h.getLocalAnalysis(alerts)
}

// getOpenAIClient creates an OpenAI client if configured
func (h *Handler) getOpenAIClient() (interface{}, error) {
	// This will be implemented when we integrate OpenAI
	// For now, return nil to fall back to local analysis
	return nil, fmt.Errorf("OpenAI not configured")
}

// getOpenAIAnalysis uses OpenAI for analysis
func (h *Handler) getOpenAIAnalysis(ctx context.Context, openaiClient interface{}, alerts []domain.Alert) (interface{}, error) {
	// Implementation will be added when OpenAI integration is complete
	return nil, fmt.Errorf("OpenAI analysis not yet implemented")
}

// getLocalAnalysis uses local ML models for analysis
func (h *Handler) getLocalAnalysis(alerts []domain.Alert) (interface{}, error) {
	// Use existing incident teller for local analysis
	teller := services.NewIncidentTeller()
	story := teller.TellStory(alerts)

	return map[string]interface{}{
		"summary":   story.Summary,
		"root_cause": story.RootCause,
		"impact":    story.Impact,
		"recommendations": map[string]interface{}{
			"immediate": story.Fix.ImmediateActions,
			"short_term": story.Fix.ShortTermActions,
			"long_term": story.Fix.LongTermActions,
		},
		"generated_at": story.GeneratedAt,
		"alert_count": len(alerts),
		"time_span": alerts[len(alerts)-1].OccurredAt.Sub(alerts[0].OccurredAt),
	}, nil
}

