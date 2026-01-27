package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"incident-teller/internal/adapters/netdata"
	"incident-teller/internal/adapters/repository"
	"incident-teller/internal/ai"
	"incident-teller/internal/api"
	"incident-teller/internal/config"
	"incident-teller/internal/database"
	"incident-teller/internal/observability"
	"incident-teller/internal/services"
)

func main() {
	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}

	// Initialize observability
	logger := observability.NewLogger(cfg.Observability)
	healthChecker := observability.NewHealthChecker(cfg.Observability.ServiceVersion)

	// Initialize database based on type
	var repo api.Repository
	switch cfg.Database.Type {
	case "sqlite":
		db, err := sql.Open("sqlite3", cfg.Database.SQLitePath)
		if err != nil {
			logger.Fatal("Failed to open SQLite database", observability.Error(err))
		}
		defer db.Close()

		sqlRepo := database.NewSQLRepository(db)
		if err := sqlRepo.Init(context.Background()); err != nil {
			logger.Fatal("Failed to initialize database", observability.Error(err))
		}
		repo = sqlRepo
	case "memory":
		repo = repository.NewInMemoryRepository()
	default:
		logger.Fatal("Unsupported database type: " + cfg.Database.Type)
	}

	// Initialize AI model (if enabled)
	var aiModel ai.AIModel
	if cfg.AI.Enabled {
		aiModel = ai.NewLocalAIModel()
	}

	// Initialize Netdata client
	netdataClient := netdata.NewClient(cfg.Netdata.BaseURL, cfg.Netdata.Hostname)

	// Register health checks
	healthChecker.RegisterCheck("database", observability.DatabaseHealthCheck(nil))
	healthChecker.RegisterCheck("netdata", observability.NetdataHealthCheck(cfg.Netdata.BaseURL))
	healthChecker.RegisterCheck("memory", observability.MemoryHealthCheck(80.0))

	// Initialize metrics
	metrics := observability.NewMetrics(cfg.Observability)

	// Initialize API handler
	handler := api.NewHandler(repo, aiModel, logger, healthChecker, metrics)

	// Setup routes with CORS middleware
	mux := handler.SetupRoutes()

	// Add CORS middleware to all routes
	corsMux := corsMiddleware(mux)

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.Server.Port),
		Handler:      corsMux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start backfill of existing alerts if any
	go backfillIncidents(context.Background(), repo, logger, cfg.Incident.CorrelationWindow)

	// Start background polling (if needed)
	if cfg.Netdata.PollInterval > 0 {
		go startPolling(context.Background(), netdataClient, repo, logger, cfg)
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting server on port " + strconv.Itoa(cfg.Server.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", observability.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", observability.Error(err))
	}

	logger.Info("Server exited")
}

// backfillIncidents correlates existing alerts into incidents
func backfillIncidents(ctx context.Context, repo api.Repository, logger observability.Logger, window time.Duration) {
	logger.Info("Checking for alerts to backfill...")
	alerts, err := repo.GetAlerts(ctx)
	if err != nil {
		logger.Error("Backfill failed to get alerts", observability.Error(err))
		return
	}

	if len(alerts) == 0 {
		return
	}

	builder := services.NewIncidentBuilder(window)
	incidents := builder.Build(alerts)

	for _, inc := range incidents {
		if err := repo.SaveIncident(ctx, inc); err != nil {
			logger.Error("Failed to backfill incident", observability.String("id", inc.ID))
		}
	}
	logger.Info("Backfill complete", observability.Int("incidents_created", len(incidents)))
}

// startPolling begins background polling for Netdata alerts
func startPolling(ctx context.Context, client *netdata.Client, repo api.Repository, logger observability.Logger, cfg *config.Config) {
	interval := cfg.Netdata.PollInterval
	logger.Info("Starting background Netdata polling",
		observability.String("interval", interval.String()))

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Background polling stopped")
			return
		case <-ticker.C:
			if err := pollOnce(ctx, client, repo, logger, cfg); err != nil {
				logger.Error("Polling error", observability.Error(err))
			}
		}
	}
}

// pollOnce performs a single polling operation
func pollOnce(ctx context.Context, client *netdata.Client, repo api.Repository, logger observability.Logger, cfg *config.Config) error {
	// Get last processed ID
	lastID, err := repo.GetLastProcessedID(ctx)
	if err != nil {
		logger.Error("Failed to get last processed ID", observability.Error(err))
		lastID = 0
	}

	// Fetch new alerts
	alerts, err := client.FetchLatest(ctx, lastID)
	if err != nil {
		return err
	}

	if len(alerts) == 0 {
		return nil
	}

	logger.Info("Fetched new alerts",
		observability.Int("count", len(alerts)),
		observability.Int64("last_id", int64(lastID)))

	// Save alerts
	var maxID uint64
	for _, alert := range alerts {
		if err := repo.SaveAlert(ctx, alert); err != nil {
			logger.Error("Failed to save alert",
				observability.Error(err),
				observability.String("alert_id", alert.ID))
			continue
		}
		if alert.ExternalID > maxID {
			maxID = alert.ExternalID
		}
	}

	// Correlate alerts into incidents
	// We use the correlation window from config
	builder := services.NewIncidentBuilder(cfg.Incident.CorrelationWindow)
	newIncidents := builder.Build(alerts)

	for _, incident := range newIncidents {
		if err := repo.SaveIncident(ctx, incident); err != nil {
			logger.Error("Failed to save incident",
				observability.Error(err),
				observability.String("incident_id", incident.ID))
		} else {
			logger.Info("Created/Updated incident from new alerts",
				observability.String("incident_id", incident.ID),
				observability.Int("alert_count", len(incident.Events)))
		}
	}

	// Update last processed ID
	if maxID > 0 {
		if err := repo.SetLastProcessedID(ctx, maxID); err != nil {
			logger.Error("Failed to update last processed ID", observability.Error(err))
		}
	}

	return nil
}

// corsMiddleware adds CORS headers to all responses
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
