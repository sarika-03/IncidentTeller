package observability

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"incident-teller/internal/config"
)

// Logger provides structured logging
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	With(fields ...Field) Logger
	WithContext(ctx context.Context) Logger
	GetLogs() []string
}

// Field represents a key-value pair for structured logging
type Field struct {
	Key   string
	Value interface{}
}

// StandardLogger provides basic structured logging
type StandardLogger struct {
	level   LogLevel
	fields  []Field
	buffer  []string
	maxSize int
}

// LogLevel represents logging level
type LogLevel int

const (
	DebugLevel LogLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// NewLogger creates a new logger instance
func NewLogger(cfg config.ObservabilityConfig) Logger {
	var level LogLevel
	switch cfg.LogLevel {
	case "debug":
		level = DebugLevel
	case "info":
		level = InfoLevel
	case "warn":
		level = WarnLevel
	case "error":
		level = ErrorLevel
	default:
		level = InfoLevel
	}

	return &StandardLogger{
		level:   level,
		buffer:  make([]string, 0),
		maxSize: 100, // Keep last 100 logs
	}
}

// GetLogs returns the buffered logs
func (l *StandardLogger) GetLogs() []string {
	return l.buffer
}

// Debug logs debug messages
func (l *StandardLogger) Debug(msg string, fields ...Field) {
	if l.level <= DebugLevel {
		l.log("DEBUG", msg, fields...)
	}
}

// Info logs info messages
func (l *StandardLogger) Info(msg string, fields ...Field) {
	if l.level <= InfoLevel {
		l.log("INFO", msg, fields...)
	}
}

// Warn logs warning messages
func (l *StandardLogger) Warn(msg string, fields ...Field) {
	if l.level <= WarnLevel {
		l.log("WARN", msg, fields...)
	}
}

// Error logs error messages
func (l *StandardLogger) Error(msg string, fields ...Field) {
	if l.level <= ErrorLevel {
		l.log("ERROR", msg, fields...)
	}
}

// Fatal logs fatal messages and exits
func (l *StandardLogger) Fatal(msg string, fields ...Field) {
	l.log("FATAL", msg, fields...)
	os.Exit(1)
}

// With adds context fields to the logger
func (l *StandardLogger) With(fields ...Field) Logger {
	newFields := make([]Field, len(l.fields)+len(fields))
	copy(newFields, l.fields)
	copy(newFields[len(l.fields):], fields)

	return &StandardLogger{
		level:  l.level,
		fields: newFields,
	}
}

// WithContext adds context to the logger
func (l *StandardLogger) WithContext(ctx context.Context) Logger {
	// Extract context information
	if ctx != nil {
		if requestID := ctx.Value("request_id"); requestID != nil {
			return l.With(Field{Key: "request_id", Value: requestID})
		}
		if traceID := ctx.Value("trace_id"); traceID != nil {
			return l.With(Field{Key: "trace_id", Value: traceID})
		}
	}
	return l
}

// log performs the actual logging
func (l *StandardLogger) log(level, msg string, fields ...Field) {
	timestamp := time.Now().UTC().Format(time.RFC3339)

	// Build log message
	logMsg := fmt.Sprintf("[%s] %s", timestamp, level)

	// Add base fields
	for _, field := range l.fields {
		logMsg += fmt.Sprintf(" %s=%v", field.Key, field.Value)
	}

	// Add message fields
	for _, field := range fields {
		logMsg += fmt.Sprintf(" %s=%v", field.Key, field.Value)
	}

	// Add message
	logMsg += fmt.Sprintf(" msg=\"%s\"", msg)

	// Add caller information for debug logs
	if level == "DEBUG" {
		_, file, line, ok := runtime.Caller(2)
		if ok {
			logMsg += fmt.Sprintf(" caller=\"%s:%d\"", file, line)
		}
	}

	// Add to buffer
	l.buffer = append(l.buffer, logMsg)
	if len(l.buffer) > l.maxSize {
		l.buffer = l.buffer[1:]
	}

	log.Println(logMsg)
}

// Metrics provides basic metrics collection
type Metrics interface {
	IncCounter(name string, labels map[string]string)
	SetGauge(name string, value float64, labels map[string]string)
	RecordHistogram(name string, value float64, labels map[string]string)
	RecordDuration(name string, duration time.Duration, labels map[string]string)
}

// StandardMetrics provides basic in-memory metrics
type StandardMetrics struct {
	counters map[string]float64
	gauges   map[string]float64
}

// NewMetrics creates a new metrics instance
func NewMetrics(cfg config.ObservabilityConfig) Metrics {
	if !cfg.EnableMetrics {
		return &NoOpMetrics{}
	}

	return &StandardMetrics{
		counters: make(map[string]float64),
		gauges:   make(map[string]float64),
	}
}

// IncCounter increments a counter metric
func (m *StandardMetrics) IncCounter(name string, labels map[string]string) {
	key := m.buildKey(name, labels)
	m.counters[key]++
}

// SetGauge sets a gauge metric
func (m *StandardMetrics) SetGauge(name string, value float64, labels map[string]string) {
	key := m.buildKey(name, labels)
	m.gauges[key] = value
}

// RecordHistogram records a histogram value
func (m *StandardMetrics) RecordHistogram(name string, value float64, labels map[string]string) {
	// For simple implementation, convert to counter
	key := m.buildKey(name+"_sum", labels)
	m.counters[key] += value

	countKey := m.buildKey(name+"_count", labels)
	m.counters[countKey]++
}

// RecordDuration records a duration as histogram
func (m *StandardMetrics) RecordDuration(name string, duration time.Duration, labels map[string]string) {
	m.RecordHistogram(name, duration.Seconds(), labels)
}

// buildKey builds a metric key with labels
func (m *StandardMetrics) buildKey(name string, labels map[string]string) string {
	if len(labels) == 0 {
		return name
	}

	key := name + "{"
	for k, v := range labels {
		key += fmt.Sprintf("%s=\"%s\",", k, v)
	}
	key = key[:len(key)-1] + "}"
	return key
}

// GetCounters returns all counters (for testing/debugging)
func (m *StandardMetrics) GetCounters() map[string]float64 {
	result := make(map[string]float64)
	for k, v := range m.counters {
		result[k] = v
	}
	return result
}

// GetGauges returns all gauges (for testing/debugging)
func (m *StandardMetrics) GetGauges() map[string]float64 {
	result := make(map[string]float64)
	for k, v := range m.gauges {
		result[k] = v
	}
	return result
}

// NoOpMetrics provides a no-op metrics implementation
type NoOpMetrics struct{}

// IncCounter no-op implementation
func (m *NoOpMetrics) IncCounter(name string, labels map[string]string) {}

// SetGauge no-op implementation
func (m *NoOpMetrics) SetGauge(name string, value float64, labels map[string]string) {}

// RecordHistogram no-op implementation
func (m *NoOpMetrics) RecordHistogram(name string, value float64, labels map[string]string) {}

// RecordDuration no-op implementation
func (m *NoOpMetrics) RecordDuration(name string, duration time.Duration, labels map[string]string) {}

// HealthChecker provides health check functionality
type HealthChecker interface {
	CheckHealth(ctx context.Context) HealthStatus
	RegisterCheck(name string, check HealthCheck)
}

// HealthCheck represents a health check function
type HealthCheck func(ctx context.Context) HealthCheckResult

// HealthCheckResult represents the result of a health check
type HealthCheckResult struct {
	Status    string                 `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Duration  time.Duration          `json:"duration"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
}

// HealthStatus represents overall system health
type HealthStatus struct {
	Status    string                       `json:"status"`
	Checks    map[string]HealthCheckResult `json:"checks"`
	Duration  time.Duration                `json:"duration"`
	Timestamp time.Time                    `json:"timestamp"`
	Version   string                       `json:"version"`
}

// StandardHealthChecker provides health check implementation
type StandardHealthChecker struct {
	checks  map[string]HealthCheck
	version string
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(version string) HealthChecker {
	return &StandardHealthChecker{
		checks:  make(map[string]HealthCheck),
		version: version,
	}
}

// CheckHealth performs all registered health checks
func (hc *StandardHealthChecker) CheckHealth(ctx context.Context) HealthStatus {
	start := time.Now()
	results := make(map[string]HealthCheckResult)

	overallStatus := "healthy"

	for name, check := range hc.checks {
		checkStart := time.Now()
		result := check(ctx)
		result.Duration = time.Since(checkStart)
		result.Timestamp = checkStart
		results[name] = result

		if result.Status != "healthy" && overallStatus != "unhealthy" {
			if result.Status == "degraded" {
				overallStatus = "degraded"
			} else {
				overallStatus = "unhealthy"
			}
		}
	}

	return HealthStatus{
		Status:    overallStatus,
		Checks:    results,
		Duration:  time.Since(start),
		Timestamp: start,
		Version:   hc.version,
	}
}

// RegisterCheck registers a new health check
func (hc *StandardHealthChecker) RegisterCheck(name string, check HealthCheck) {
	hc.checks[name] = check
}

// Common health checks

// DatabaseHealthCheck creates a database health check
func DatabaseHealthCheck(db interface{}) HealthCheck {
	return func(ctx context.Context) HealthCheckResult {
		type pinger interface {
			PingContext(context.Context) error
		}

		if p, ok := db.(pinger); ok {
			if err := p.PingContext(ctx); err != nil {
				return HealthCheckResult{
					Status:  "unhealthy",
					Message: fmt.Sprintf("Database ping failed: %v", err),
					Details: map[string]interface{}{"error": err.Error()},
				}
			}
		} else if db != nil {
			// If it's not a direct sql.DB but some other interface we don't know
			// we just report it as unknown but potentially OK if passed
			return HealthCheckResult{
				Status:  "healthy",
				Message: "Database connection exists",
				Details: map[string]interface{}{"type": "unknown_pinger"},
			}
		} else {
			return HealthCheckResult{
				Status:  "unhealthy",
				Message: "Database connection is nil",
			}
		}

		return HealthCheckResult{
			Status:  "healthy",
			Message: "Database connection OK",
			Details: map[string]interface{}{
				"type": "database",
			},
		}
	}
}

// NetdataHealthCheck creates a Netdata health check
func NetdataHealthCheck(baseURL string) HealthCheck {
	return func(ctx context.Context) HealthCheckResult {
		client := &http.Client{
			Timeout: 5 * time.Second,
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/api/v1/info", nil)
		if err != nil {
			return HealthCheckResult{
				Status:  "unhealthy",
				Message: fmt.Sprintf("Failed to create request: %v", err),
			}
		}

		resp, err := client.Do(req)
		if err != nil {
			return HealthCheckResult{
				Status:  "unhealthy",
				Message: fmt.Sprintf("Netdata API unreachable: %v", err),
				Details: map[string]interface{}{"url": baseURL, "error": err.Error()},
			}
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return HealthCheckResult{
				Status:  "unhealthy",
				Message: fmt.Sprintf("Netdata API returned status %d", resp.StatusCode),
				Details: map[string]interface{}{"url": baseURL, "status": resp.StatusCode},
			}
		}

		return HealthCheckResult{
			Status:  "healthy",
			Message: "Netdata API reachable",
			Details: map[string]interface{}{
				"url": baseURL,
			},
		}
	}
}

// MemoryHealthCheck creates a memory health check
func MemoryHealthCheck(thresholdPercent float64) HealthCheck {
	return func(ctx context.Context) HealthCheckResult {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		memoryUsedMB := float64(m.Alloc) / 1024 / 1024
		memoryTotalMB := float64(m.Sys) / 1024 / 1024
		percentUsed := (memoryUsedMB / memoryTotalMB) * 100

		status := "healthy"
		if percentUsed > thresholdPercent {
			status = "degraded"
		}
		if percentUsed > thresholdPercent*1.5 {
			status = "unhealthy"
		}

		return HealthCheckResult{
			Status:  status,
			Message: fmt.Sprintf("Memory usage: %.1f%% (%.1fMB/%.1fMB)", percentUsed, memoryUsedMB, memoryTotalMB),
			Details: map[string]interface{}{
				"used_mb":      memoryUsedMB,
				"total_mb":     memoryTotalMB,
				"percent_used": percentUsed,
				"threshold":    thresholdPercent,
				"alloc_bytes":  m.Alloc,
				"sys_bytes":    m.Sys,
			},
		}
	}
}

// Utility functions

// String creates a string field
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int creates an int field
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 creates an int64 field
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Float64 creates a float64 field
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// Bool creates a bool field
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Duration creates a duration field
func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}

// Time creates a time field
func Time(key string, value time.Time) Field {
	return Field{Key: key, Value: value}
}

// Error creates an error field
func Error(err error) Field {
	return Field{Key: "error", Value: err.Error()}
}

// Any creates a field with any type
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}
