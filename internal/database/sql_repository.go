package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"incident-teller/internal/domain"
)

// SQLRepository provides persistent storage using SQL databases
type SQLRepository struct {
	db *sql.DB
}

// NewSQLRepository creates a new SQL repository
func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

// Init initializes database tables
func (r *SQLRepository) Init(ctx context.Context) error {
	// Create tables first
	queries := []string{
		`CREATE TABLE IF NOT EXISTS alerts (
			id TEXT PRIMARY KEY,
			external_id INTEGER NOT NULL,
			host TEXT NOT NULL,
			chart TEXT NOT NULL,
			family TEXT NOT NULL,
			name TEXT NOT NULL,
			status TEXT NOT NULL,
			old_status TEXT NOT NULL,
			value REAL NOT NULL,
			occurred_at TIMESTAMP NOT NULL,
			description TEXT,
			resource_type TEXT NOT NULL,
			labels TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS incidents (
			id TEXT PRIMARY KEY,
			title TEXT NOT NULL,
			status TEXT NOT NULL,
			started_at TIMESTAMP NOT NULL,
			resolved_at TIMESTAMP,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS incident_alerts (
			incident_id TEXT NOT NULL,
			alert_id TEXT NOT NULL,
			sequence_order INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (incident_id, alert_id),
			FOREIGN KEY (incident_id) REFERENCES incidents(id) ON DELETE CASCADE,
			FOREIGN KEY (alert_id) REFERENCES alerts(id) ON DELETE CASCADE
		)`,
		`CREATE TABLE IF NOT EXISTS metadata (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,

		// Create indexes separately
		`CREATE INDEX IF NOT EXISTS idx_alerts_external_id ON alerts(external_id)`,
		`CREATE INDEX IF NOT EXISTS idx_alerts_occurred_at ON alerts(occurred_at)`,
		`CREATE INDEX IF NOT EXISTS idx_alerts_host ON alerts(host)`,
		`CREATE INDEX IF NOT EXISTS idx_alerts_resource_type ON alerts(resource_type)`,
		`CREATE INDEX IF NOT EXISTS idx_incidents_status ON incidents(status)`,
		`CREATE INDEX IF NOT EXISTS idx_incidents_started_at ON incidents(started_at)`,
		`CREATE INDEX IF NOT EXISTS idx_incidents_resolved_at ON incidents(resolved_at)`,
		`CREATE INDEX IF NOT EXISTS idx_incident_alerts_incident_id ON incident_alerts(incident_id)`,
		`CREATE INDEX IF NOT EXISTS idx_incident_alerts_alert_id ON incident_alerts(alert_id)`,
		`CREATE INDEX IF NOT EXISTS idx_incident_alerts_sequence_order ON incident_alerts(sequence_order)`,
	}

	for _, query := range queries {
		if _, err := r.db.ExecContext(ctx, query); err != nil {
			return fmt.Errorf("failed to execute migration query: %w", err)
		}
	}

	return nil
}

// SaveAlert stores an alert in the database
func (r *SQLRepository) SaveAlert(ctx context.Context, alert domain.Alert) error {
	labelsJSON, err := json.Marshal(alert.Labels)
	if err != nil {
		return fmt.Errorf("failed to marshal labels: %w", err)
	}

	query := `
		INSERT INTO alerts (
			id, external_id, host, chart, family, name, status, old_status,
			value, occurred_at, description, resource_type, labels
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			status = excluded.status,
			old_status = excluded.old_status,
			value = excluded.value,
			occurred_at = excluded.occurred_at,
			description = excluded.description,
			labels = excluded.labels
	`

	_, err = r.db.ExecContext(ctx, query,
		alert.ID, alert.ExternalID, alert.Host, alert.Chart, alert.Family,
		alert.Name, string(alert.Status), string(alert.OldStatus),
		alert.Value, alert.OccurredAt, alert.Description,
		string(alert.ResourceType), string(labelsJSON),
	)

	return err
}

// GetIncidents retrieves incidents from the database
func (r *SQLRepository) GetIncidents(ctx context.Context) ([]domain.Incident, error) {
	query := `
		SELECT id, title, status, started_at, resolved_at
		FROM incidents
		ORDER BY started_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query incidents: %w", err)
	}
	defer rows.Close()

	var incidents []domain.Incident
	for rows.Next() {
		var incident domain.Incident
		var resolvedAt sql.NullTime

		err := rows.Scan(
			&incident.ID, &incident.Title, &incident.Status,
			&incident.StartedAt, &resolvedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan incident: %w", err)
		}

		if resolvedAt.Valid {
			incident.ResolvedAt = &resolvedAt.Time
		}

		// Load associated alerts
		alerts, err := r.getIncidentAlerts(ctx, incident.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get incident alerts: %w", err)
		}

		incident.Events = alerts
		incidents = append(incidents, incident)
	}

	return incidents, rows.Err()
}

// SaveIncident stores an incident in the database
func (r *SQLRepository) SaveIncident(ctx context.Context, incident domain.Incident) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO incidents (id, title, status, started_at, resolved_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			title = excluded.title,
			status = excluded.status,
			resolved_at = excluded.resolved_at,
			updated_at = CURRENT_TIMESTAMP
	`

	var resolvedAt interface{}
	if incident.ResolvedAt != nil {
		resolvedAt = *incident.ResolvedAt
	}

	_, err = tx.ExecContext(ctx, query,
		incident.ID, incident.Title, string(incident.Status),
		incident.StartedAt, resolvedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to upsert incident: %w", err)
	}

	// Delete existing incident_alerts relations
	_, err = tx.ExecContext(ctx, "DELETE FROM incident_alerts WHERE incident_id = ?", incident.ID)
	if err != nil {
		return fmt.Errorf("failed to delete incident alerts: %w", err)
	}

	// Insert incident_alerts relations
	for i, alert := range incident.Events {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO incident_alerts (incident_id, alert_id, sequence_order)
			VALUES (?, ?, ?)
		`, incident.ID, alert.ID, i)
		if err != nil {
			return fmt.Errorf("failed to insert incident alert: %w", err)
		}
	}

	return tx.Commit()
}

// GetLastProcessedID returns the last processed alert ID
func (r *SQLRepository) GetLastProcessedID(ctx context.Context) (uint64, error) {
	var value string
	query := "SELECT value FROM metadata WHERE key = 'last_processed_id'"

	err := r.db.QueryRowContext(ctx, query).Scan(&value)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get last processed ID: %w", err)
	}

	var id uint64
	_, err = fmt.Sscanf(value, "%d", &id)
	if err != nil {
		return 0, fmt.Errorf("failed to parse last processed ID: %w", err)
	}

	return id, nil
}

// SetLastProcessedID updates the last processed alert ID
func (r *SQLRepository) SetLastProcessedID(ctx context.Context, id uint64) error {
	query := `
		INSERT INTO metadata (key, value) VALUES ('last_processed_id', ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP
	`

	_, err := r.db.ExecContext(ctx, query, fmt.Sprintf("%d", id))
	return err
}

// GetAlerts retrieves alerts from the database
func (r *SQLRepository) GetAlerts(ctx context.Context) ([]domain.Alert, error) {
	query := `
		SELECT id, external_id, host, chart, family, name, status, old_status,
			   value, occurred_at, description, resource_type, labels
		FROM alerts
		ORDER BY occurred_at DESC
		LIMIT 1000
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query alerts: %w", err)
	}
	defer rows.Close()

	var alerts []domain.Alert
	for rows.Next() {
		var alert domain.Alert
		var labelsJSON string
		var description sql.NullString

		err := rows.Scan(
			&alert.ID, &alert.ExternalID, &alert.Host, &alert.Chart,
			&alert.Family, &alert.Name, &alert.Status, &alert.OldStatus,
			&alert.Value, &alert.OccurredAt, &description,
			&alert.ResourceType, &labelsJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}

		if description.Valid {
			alert.Description = description.String
		}

		if labelsJSON != "" {
			if err := json.Unmarshal([]byte(labelsJSON), &alert.Labels); err != nil {
				return nil, fmt.Errorf("failed to unmarshal labels: %w", err)
			}
		}

		alerts = append(alerts, alert)
	}

	return alerts, rows.Err()
}

// getIncidentAlerts retrieves alerts for a specific incident
func (r *SQLRepository) getIncidentAlerts(ctx context.Context, incidentID string) ([]domain.Alert, error) {
	query := `
		SELECT a.id, a.external_id, a.host, a.chart, a.family, a.name, 
			   a.status, a.old_status, a.value, a.occurred_at, a.description, 
			   a.resource_type, a.labels
		FROM alerts a
		JOIN incident_alerts ia ON a.id = ia.alert_id
		WHERE ia.incident_id = ?
		ORDER BY ia.sequence_order
	`

	rows, err := r.db.QueryContext(ctx, query, incidentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query incident alerts: %w", err)
	}
	defer rows.Close()

	var alerts []domain.Alert
	for rows.Next() {
		var alert domain.Alert
		var labelsJSON string
		var description sql.NullString

		err := rows.Scan(
			&alert.ID, &alert.ExternalID, &alert.Host, &alert.Chart,
			&alert.Family, &alert.Name, &alert.Status, &alert.OldStatus,
			&alert.Value, &alert.OccurredAt, &description,
			&alert.ResourceType, &labelsJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}

		if description.Valid {
			alert.Description = description.String
		}

		if labelsJSON != "" {
			if err := json.Unmarshal([]byte(labelsJSON), &alert.Labels); err != nil {
				return nil, fmt.Errorf("failed to unmarshal labels: %w", err)
			}
		}

		alerts = append(alerts, alert)
	}

	return alerts, rows.Err()
}

// Close closes the database connection
func (r *SQLRepository) Close() error {
	return r.db.Close()
}

// Stats returns repository statistics
func (r *SQLRepository) Stats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Count alerts
	var alertCount int
	err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM alerts").Scan(&alertCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count alerts: %w", err)
	}
	stats["total_alerts"] = alertCount

	// Count incidents
	var incidentCount int
	err = r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM incidents").Scan(&incidentCount)
	if err != nil {
		return nil, fmt.Errorf("failed to count incidents: %w", err)
	}
	stats["total_incidents"] = incidentCount

	// Get last processed ID
	lastID, err := r.GetLastProcessedID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get last processed ID: %w", err)
	}
	stats["last_processed_id"] = lastID

	// Get database size (for SQLite) - simple heuristic
	if len(stats) > 0 {
		// Just add a placeholder for now since we can't easily detect driver type
		stats["db_size_bytes"] = "unknown"
	}

	return stats, nil
}

// CreateIncidentFromAlerts creates an incident from a list of alerts
func (r *SQLRepository) CreateIncidentFromAlerts(ctx context.Context, alerts []domain.Alert) (*domain.Incident, error) {
	if len(alerts) == 0 {
		return nil, fmt.Errorf("no alerts provided")
	}

	// Generate incident ID
	incidentID := fmt.Sprintf("incident-%s-%d",
		alerts[0].Host, alerts[0].OccurredAt.Unix())

	// Create incident title from first alert
	title := fmt.Sprintf("%s on %s", alerts[0].Name, alerts[0].Host)

	// Determine incident status
	status := domain.StatusCritical
	for _, alert := range alerts {
		if alert.Status == domain.StatusWarning {
			status = domain.StatusWarning
		}
	}

	incident := domain.Incident{
		ID:        incidentID,
		Title:     title,
		Status:    status,
		StartedAt: alerts[0].OccurredAt,
		Events:    alerts,
	}

	// Save incident
	err := r.SaveIncident(ctx, incident)
	if err != nil {
		return nil, fmt.Errorf("failed to save incident: %w", err)
	}

	return &incident, nil
}

// GetIncidentsByTimeRange retrieves incidents within a time range
func (r *SQLRepository) GetIncidentsByTimeRange(ctx context.Context, start, end time.Time) ([]domain.Incident, error) {
	query := `
		SELECT id, title, status, started_at, resolved_at
		FROM incidents
		WHERE started_at >= ? AND started_at <= ?
		ORDER BY started_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, start, end)
	if err != nil {
		return nil, fmt.Errorf("failed to query incidents by time range: %w", err)
	}
	defer rows.Close()

	var incidents []domain.Incident
	for rows.Next() {
		var incident domain.Incident
		var resolvedAt sql.NullTime

		err := rows.Scan(
			&incident.ID, &incident.Title, &incident.Status,
			&incident.StartedAt, &resolvedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan incident: %w", err)
		}

		if resolvedAt.Valid {
			incident.ResolvedAt = &resolvedAt.Time
		}

		// Load associated alerts
		alerts, err := r.getIncidentAlerts(ctx, incident.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get incident alerts: %w", err)
		}

		incident.Events = alerts
		incidents = append(incidents, incident)
	}

	return incidents, rows.Err()
}

// DeleteOldAlerts removes alerts older than the specified duration
func (r *SQLRepository) DeleteOldAlerts(ctx context.Context, olderThan time.Duration) error {
	query := "DELETE FROM alerts WHERE occurred_at < ?"

	_, err := r.db.ExecContext(ctx, query, time.Now().Add(-olderThan))
	return err
}

// PingContext checks database connectivity
func (r *SQLRepository) PingContext(ctx context.Context) error {
	return r.db.PingContext(ctx)
}
