package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v6"
	"gopkg.in/yaml.v3"
)

// Config holds all configuration for IncidentTeller
type Config struct {
	Server        ServerConfig        `yaml:"server" envPrefix:"SERVER_"`
	Netdata       NetdataConfig       `yaml:"netdata" envPrefix:"NETDATA_"`
	AI            AIConfig            `yaml:"ai" envPrefix:"AI_"`
	Database      DatabaseConfig      `yaml:"database" envPrefix:"DB_"`
	Observability ObservabilityConfig `yaml:"observability" envPrefix:"OBSERVABILITY_"`
	Incident      IncidentConfig      `yaml:"incident" envPrefix:"INCIDENT_"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host         string        `yaml:"host" env:"HOST" envDefault:"0.0.0.0"`
	Port         int           `yaml:"port" env:"PORT" envDefault:"8080"`
	ReadTimeout  time.Duration `yaml:"read_timeout" env:"READ_TIMEOUT" envDefault:"30s"`
	WriteTimeout time.Duration `yaml:"write_timeout" env:"WRITE_TIMEOUT" envDefault:"30s"`
	IdleTimeout  time.Duration `yaml:"idle_timeout" env:"IDLE_TIMEOUT" envDefault:"120s"`
}

// NetdataConfig holds Netdata API configuration
type NetdataConfig struct {
	BaseURL      string        `yaml:"base_url" env:"BASE_URL" envDefault:"http://localhost:19999"`
	Timeout      time.Duration `yaml:"timeout" env:"TIMEOUT" envDefault:"30s"`
	RetryCount   int           `yaml:"retry_count" env:"RETRY_COUNT" envDefault:"3"`
	RetryDelay   time.Duration `yaml:"retry_delay" env:"RETRY_DELAY" envDefault:"1s"`
	PollInterval time.Duration `yaml:"poll_interval" env:"POLL_INTERVAL" envDefault:"10s"`
	Hostname     string        `yaml:"hostname" env:"HOSTNAME" envDefault:"localhost"`
	BatchSize    int           `yaml:"batch_size" env:"BATCH_SIZE" envDefault:"100"`
}

// AIConfig holds AI/ML configuration
type AIConfig struct {
	Enabled             bool          `yaml:"enabled" env:"ENABLED" envDefault:"true"`
	ModelType           string        `yaml:"model_type" env:"MODEL_TYPE" envDefault:"local"`
	APIToken            string        `yaml:"api_token" env:"API_TOKEN"`
	APIEndpoint         string        `yaml:"api_endpoint" env:"API_ENDPOINT"`
	ConfidenceThreshold float64       `yaml:"confidence_threshold" env:"CONFIDENCE_THRESHOLD" envDefault:"0.7"`
	MaxPredictions      int           `yaml:"max_predictions" env:"MAX_PREDICTIONS" envDefault:"5"`
	PredictionTimeout   time.Duration `yaml:"prediction_timeout" env:"PREDICTION_TIMEOUT" envDefault:"10s"`
	EnableLearning      bool          `yaml:"enable_learning" env:"ENABLE_LEARNING" envDefault:"false"`
	ModelPath           string        `yaml:"model_path" env:"MODEL_PATH" envDefault:"./models"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Type            string        `yaml:"type" env:"TYPE" envDefault:"memory"`
	Host            string        `yaml:"host" env:"HOST" envDefault:"localhost"`
	Port            int           `yaml:"port" env:"PORT" envDefault:"5432"`
	Database        string        `yaml:"database" env:"DATABASE" envDefault:"incident_teller"`
	Username        string        `yaml:"username" env:"USERNAME" envDefault:"incident_teller"`
	Password        string        `yaml:"password" env:"PASSWORD" envDefault:""`
	SSLMode         string        `yaml:"ssl_mode" env:"SSL_MODE" envDefault:"disable"`
	MaxConnections  int           `yaml:"max_connections" env:"MAX_CONNECTIONS" envDefault:"10"`
	MaxIdleConns    int           `yaml:"max_idle_conns" env:"MAX_IDLE_CONNS" envDefault:"5"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" env:"CONN_MAX_LIFETIME" envDefault:"1h"`
	SQLitePath      string        `yaml:"sqlite_path" env:"SQLITE_PATH" envDefault:"./incident_teller.db"`
}

// ObservabilityConfig holds observability configuration
type ObservabilityConfig struct {
	LogLevel        string            `yaml:"log_level" env:"LOG_LEVEL" envDefault:"info"`
	LogFormat       string            `yaml:"log_format" env:"LOG_FORMAT" envDefault:"json"`
	EnableMetrics   bool              `yaml:"enable_metrics" env:"ENABLE_METRICS" envDefault:"true"`
	MetricsPort     int               `yaml:"metrics_port" env:"METRICS_PORT" envDefault:"9090"`
	EnableTracing   bool              `yaml:"enable_tracing" env:"ENABLE_TRACING" envDefault:"false"`
	TracingEndpoint string            `yaml:"tracing_endpoint" env:"TRACING_ENDPOINT"`
	ServiceName     string            `yaml:"service_name" env:"SERVICE_NAME" envDefault:"incident-teller"`
	ServiceVersion  string            `yaml:"service_version" env:"SERVICE_VERSION" envDefault:"1.0.0"`
	Tags            map[string]string `yaml:"tags" env:"TAGS"`
}

// IncidentConfig holds incident processing configuration
type IncidentConfig struct {
	CorrelationWindow time.Duration `yaml:"correlation_window" env:"CORRELATION_WINDOW" envDefault:"15m"`
	IncidentTimeout   time.Duration `yaml:"incident_timeout" env:"INCIDENT_TIMEOUT" envDefault:"24h"`
	MaxIncidents      int           `yaml:"max_incidents" env:"MAX_INCIDENTS" envDefault:"1000"`
	EnableAutoResolve bool          `yaml:"enable_auto_resolve" env:"ENABLE_AUTO_RESOLVE" envDefault:"true"`
	ResolveThreshold  time.Duration `yaml:"resolve_threshold" env:"RESOLVE_THRESHOLD" envDefault:"30m"`
	EnableAlertDedup  bool          `yaml:"enable_alert_dedup" env:"ENABLE_ALERT_DEDUP" envDefault:"true"`
	DedupWindow       time.Duration `yaml:"dedup_window" env:"DEDUP_WINDOW" envDefault:"5m"`
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	// Start with defaults
	cfg := &Config{}

	// Load from file if provided
	if configPath != "" {
		if err := loadFromFile(cfg, configPath); err != nil {
			return nil, fmt.Errorf("failed to load config from file: %w", err)
		}
	}

	// Override with environment variables
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// loadFromFile loads configuration from YAML file
func loadFromFile(cfg *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Validate server config
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535")
	}

	// Validate netdata config
	if c.Netdata.BaseURL == "" {
		return fmt.Errorf("netdata base URL is required")
	}

	if _, err := time.ParseDuration(c.Netdata.Timeout.String()); err != nil {
		return fmt.Errorf("invalid netdata timeout format")
	}

	// Validate AI config
	if c.AI.Enabled {
		if c.AI.ModelType == "" {
			return fmt.Errorf("AI model type is required when AI is enabled")
		}

		if c.AI.ConfidenceThreshold < 0 || c.AI.ConfidenceThreshold > 1 {
			return fmt.Errorf("AI confidence threshold must be between 0 and 1")
		}
	}

	// Validate database config
	if c.Database.Type == "" {
		return fmt.Errorf("database type is required")
	}

	switch c.Database.Type {
	case "postgres", "postgresql":
		if c.Database.Host == "" {
			return fmt.Errorf("database host is required for PostgreSQL")
		}
		if c.Database.Port <= 0 || c.Database.Port > 65535 {
			return fmt.Errorf("database port must be between 1 and 65535")
		}
		if c.Database.Database == "" {
			return fmt.Errorf("database name is required")
		}
	case "mysql":
		if c.Database.Host == "" {
			return fmt.Errorf("database host is required for MySQL")
		}
		if c.Database.Port <= 0 || c.Database.Port > 65535 {
			return fmt.Errorf("database port must be between 1 and 65535")
		}
	case "sqlite":
		if c.Database.SQLitePath == "" {
			return fmt.Errorf("SQLite path is required")
		}
	case "memory":
		// No validation needed for in-memory
	default:
		return fmt.Errorf("unsupported database type: %s", c.Database.Type)
	}

	// Validate observability config
	validLogLevels := []string{"debug", "info", "warn", "error"}
	found := false
	for _, level := range validLogLevels {
		if c.Observability.LogLevel == level {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("invalid log level: %s", c.Observability.LogLevel)
	}

	if c.Observability.MetricsPort <= 0 || c.Observability.MetricsPort > 65535 {
		return fmt.Errorf("metrics port must be between 1 and 65535")
	}

	// Validate incident config
	if c.Incident.MaxIncidents <= 0 {
		return fmt.Errorf("max incidents must be positive")
	}

	return nil
}

// GetDSN returns database connection string
func (c *DatabaseConfig) GetDSN() string {
	switch c.Type {
	case "postgres", "postgresql":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			c.Host, c.Port, c.Username, c.Password, c.Database, c.SSLMode)
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			c.Username, c.Password, c.Host, c.Port, c.Database)
	case "sqlite":
		return c.SQLitePath
	default:
		return ""
	}
}

// IsProduction checks if running in production mode
func (c *Config) IsProduction() bool {
	return strings.ToLower(os.Getenv("ENV")) == "production" ||
		strings.ToLower(os.Getenv("ENVIRONMENT")) == "production"
}

// IsDevelopment checks if running in development mode
func (c *Config) IsDevelopment() bool {
	env := strings.ToLower(os.Getenv("ENV"))
	environment := strings.ToLower(os.Getenv("ENVIRONMENT"))
	return env == "development" || environment == "development" ||
		env == "dev" || environment == "dev"
}

// GetEnvString returns environment variable string with default
func GetEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvInt returns environment variable int with default
func GetEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

// GetEnvBool returns environment variable bool with default
func GetEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

// GetEnvDuration returns environment variable duration with default
func GetEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// GetEnvStringSlice returns environment variable string slice with default
func GetEnvStringSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		return strings.Split(value, ",")
	}
	return defaultValue
}
