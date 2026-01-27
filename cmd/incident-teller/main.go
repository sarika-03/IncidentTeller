package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"incident-teller/internal/adapters/netdata"
	"incident-teller/internal/adapters/repository"
	"incident-teller/internal/ai"
	"incident-teller/internal/api"
	"incident-teller/internal/config"
	"incident-teller/internal/database"
	"incident-teller/internal/observability"
	"incident-teller/internal/ports"
	"incident-teller/internal/services"
)

func main() {
	// Parse command-line flags
	configPath := flag.String("config", "", "Path to configuration file")
	version := flag.Bool("version", false, "Show version information")
	flag.Parse()

	if *version {
		fmt.Println("IncidentTeller v1.0.0")
		os.Exit(0)
	}

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize observability
	logger := observability.NewLogger(cfg.Observability)
	metrics := observability.NewMetrics(cfg.Observability)
	healthChecker := observability.NewHealthChecker("1.0.0")

	logger.Info("Starting IncidentTeller",
		observability.String("version", "1.0.0"),
		observability.String("config_source", func() string {
			if *configPath != "" {
				return "file"
			}
			return "env"
		}()),
	)

	// Initialize database
	var db *sql.DB
	var repo api.Repository

	switch cfg.Database.Type {
	case "postgres", "postgresql":
		db, err = sql.Open("postgres", cfg.Database.GetDSN())
	case "mysql":
		db, err = sql.Open("mysql", cfg.Database.GetDSN())
	case "sqlite":
		db, err = sql.Open("sqlite3", cfg.Database.GetDSN())
	case "memory":
		memoryRepo := repository.NewInMemoryRepository()
		repo = memoryRepo
		logger.Info("Using in-memory repository")
	default:
		logger.Fatal("Unsupported database type", observability.String("type", cfg.Database.Type))
	}

	if err != nil {
		logger.Fatal("Failed to connect to database", observability.Error(err))
	}

	if db != nil {
		defer db.Close()

		// Configure connection pool
		db.SetMaxOpenConns(cfg.Database.MaxConnections)
		db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
		db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

		// Initialize SQL repository
		sqlRepo := database.NewSQLRepository(db)
		initCtx, initCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer initCancel()

		if err := sqlRepo.Init(initCtx); err != nil {
			logger.Fatal("Failed to initialize database", observability.Error(err))
		}

		repo = sqlRepo
		logger.Info("Database initialized",
			observability.String("type", cfg.Database.Type))
	}

	// Register health checks
	healthChecker.RegisterCheck("database", observability.DatabaseHealthCheck(repo))
	healthChecker.RegisterCheck("netdata", observability.NetdataHealthCheck(cfg.Netdata.BaseURL))
	healthChecker.RegisterCheck("memory", observability.MemoryHealthCheck(80.0))

	// Initialize Netdata client (supports both local and cloud)
	var netdataClient ports.AlertSource

	if cfg.Netdata.CloudEnabled {
		logger.Info("Using Netdata Cloud API",
			observability.String("space", cfg.Netdata.CloudSpace))

		netdataClient = netdata.NewCloudClient(
			cfg.Netdata.CloudToken,
			cfg.Netdata.CloudSpace,
			cfg.Netdata.CloudRooms...,
		)
	} else {
		logger.Info("Using Local Netdata API",
			observability.String("url", cfg.Netdata.BaseURL))

		netdataClient = netdata.NewClient(
			cfg.Netdata.BaseURL,
			cfg.Netdata.Hostname,
		)
	}

	// Initialize AI model
	var aiModel ai.AIModel
	if cfg.AI.Enabled {
		aiModel = ai.NewLocalAIModel()
		logger.Info("AI model enabled",
			observability.String("type", cfg.AI.ModelType),
			observability.Float64("confidence_threshold", cfg.AI.ConfidenceThreshold))
	} else {
		logger.Info("AI model disabled")
	}

	// Initialize analyzers
	incidentAnalyzer := services.NewIncidentAnalyzer()

	// Initialize enhanced poller
	poller := services.NewRealTimePoller(
		netdataClient,
		repo,
		incidentAnalyzer,
		cfg.Netdata.PollInterval,
	)

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start metrics server if enabled
	if cfg.Observability.EnableMetrics {
		go func() {
			metricsAddr := fmt.Sprintf(":%d", cfg.Observability.MetricsPort)
			logger.Info("Starting metrics server", observability.String("addr", metricsAddr))

			// Simple metrics endpoint
			http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "text/plain")
				fmt.Fprintf(w, "# IncidentTeller Metrics\n")
				fmt.Fprintf(w, "incident_teller_uptime_seconds %f\n", time.Since(time.Now()).Seconds())
				fmt.Fprintf(w, "incident_teller_build_info{version=\"1.0.0\"} 1\n")
			})

			if err := http.ListenAndServe(metricsAddr, nil); err != nil {
				logger.Error("Metrics server failed", observability.Error(err))
			}
		}()
	}

	// Initialize API handlers
	apiHandler := api.NewHandler(repo, aiModel, logger, healthChecker, metrics)

	// Start API server
	go func() {
		apiAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
		logger.Info("Starting API server", observability.String("addr", apiAddr))

		mux := apiHandler.SetupRoutes()

		if err := http.ListenAndServe(apiAddr, mux); err != nil {
			logger.Error("API server failed", observability.Error(err))
		}
	}()

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Start poller in background
	go func() {
		logger.Info("Starting alert poller",
			observability.String("interval", cfg.Netdata.PollInterval.String()))

		if err := poller.Start(ctx); err != nil && err != context.Canceled {
			logger.Error("Poller error", observability.Error(err))
		}
	}()

	// Monitor events and perform AI analysis
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case alerts := <-poller.Events():
				logger.Info("Received alerts for analysis",
					observability.Int("count", len(alerts)))

				metrics.RecordDuration("alerts_received_duration", time.Since(time.Now()), nil)

				// Perform comprehensive analysis
				timeline := incidentAnalyzer.AnalyzeIncident(alerts)

				// Generate AI-powered insights if enabled
				if cfg.AI.Enabled && aiModel != nil {
					aiCtx, aiCancel := context.WithTimeout(ctx, cfg.AI.PredictionTimeout)
					defer aiCancel()

					rootCause, err := aiModel.PredictRootCause(aiCtx, alerts)
					if err != nil {
						logger.Warn("AI prediction failed", observability.Error(err))
					} else {
						logger.Info("AI root cause prediction",
							observability.Float64("confidence", rootCause.Confidence),
							observability.String("pattern_type", rootCause.PatternType))

						metrics.RecordHistogram("ai_predictions_total", 1, map[string]string{
							"type": "root_cause",
						})
					}

					blastRadius, err := aiModel.PredictBlastRadius(aiCtx, alerts)
					if err != nil {
						logger.Warn("AI blast radius prediction failed", observability.Error(err))
					} else {
						logger.Info("AI blast radius prediction",
							observability.Float64("impact_score", blastRadius.ImpactScore),
							observability.String("risk_level", blastRadius.RiskLevel))

						metrics.RecordHistogram("ai_predictions_total", 1, map[string]string{
							"type": "blast_radius",
						})
					}
				}

				// Generate summary
				summary := incidentAnalyzer.GenerateIncidentSummary(timeline)
				logger.Info("Incident analysis completed",
					observability.String("summary", summary))

				metrics.RecordHistogram("incidents_analyzed_total", 1, nil)
			}
		}
	}()

	logger.Info("IncidentTeller started successfully",
		observability.String("mode", func() string {
			if cfg.IsProduction() {
				return "production"
			}
			return "development"
		}()))

	// Wait for shutdown signal
	<-sigChan
	logger.Info("Shutdown signal received")

	// Cancel context and wait for graceful shutdown
	cancel()
	time.Sleep(2 * time.Second)

	// Print final statistics
	if sqlRepo, ok := repo.(*database.SQLRepository); ok {
		stats, err := sqlRepo.Stats(context.Background())
		if err != nil {
			logger.Error("Failed to get final statistics", observability.Error(err))
		} else {
			logger.Info("Final statistics", observability.Any("stats", stats))
		}
	}

	logger.Info("IncidentTeller stopped")
}
