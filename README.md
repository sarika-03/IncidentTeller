# IncidentTeller - AI-Powered SRE Incident Analysis

**IncidentTeller** is a comprehensive incident analysis engine that transforms raw alert data into actionable intelligence for SRE teams. It uses advanced root cause analysis, blast radius detection, and AI-powered insights to help on-call engineers quickly understand and resolve production incidents.

## 🚀 New Features in v1.0.0

### 🤖 Real AI/ML Integration
- **Local ML Models**: Built-in machine learning for root cause prediction
- **Pattern Recognition**: Identifies temporal patterns (burst, cascade, gradual, spike)
- **Confidence Scoring**: AI-driven confidence levels with explainable reasoning
- **Predictive Analytics**: Predicts blast radius and incident duration

### ⚙️ Production-Ready Configuration
- **Environment-based Config**: Support for YAML files and environment variables
- **Multiple Databases**: PostgreSQL, MySQL, SQLite, or in-memory storage
- **Health Checks**: Built-in health endpoints for monitoring
- **Metrics & Observability**: Structured logging and metrics collection

### 🏗️ Enhanced Architecture
- **Clean Architecture**: Proper separation of concerns with hexagonal design
- **Database Persistence**: Full SQL storage with migrations
- **Graceful Shutdown**: Proper signal handling and cleanup
- **Error Resilience**: Retry mechanisms and circuit breakers

## Features

### 🎯 AI-Powered Root Cause Analysis
- **Confidence Scores (0-100)** for each potential root cause
- **ML Feature Extraction**: Temporal, resource, and severity patterns
- **Evidence-based Reasoning**: Explainable AI showing why one cause is more likely
- **Alternative Causes**: Multiple root cause candidates with confidence gaps
- **Pattern Types**: Burst, cascade, gradual, spike, and progressive patterns

### 💥 Intelligent Blast Radius Analysis  
- **Component Classification**: Direct, indirect, and unaffected components
- **Impact Scoring**: AI-driven 0-100 impact assessment
- **Cascade Detection**: Multi-level failure propagation analysis
- **Business Impact**: User-facing impact assessment
- **Recovery Prediction**: Estimated time to resolution

### 🔧 Context-Aware Fix Recommendations
- **Resource-Specific Playbooks**: Memory, CPU, disk, network, process fixes
- **Cascade Mitigation**: Specific actions for cascading failures
- **Complexity Assessment**: Simple, moderate, or complex incident classification
- **Time Estimates**: Predicted resolution time based on incident characteristics

### 📖 AI-Enhanced Incident Storytelling
- **Narrative Generation**: Converts technical data into human-readable stories
- **Timeline Causality**: Cause → effect sequences with clear relationships
- **Impact Communication**: Business-appropriate impact descriptions
- **Action Planning**: Prioritized remediation steps

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Netdata API  │───▶│  Alert Poller   │───▶│   AI Models     │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │                        │
                                ▼                        ▼
                       ┌──────────────────┐    ┌─────────────────┐
                       │   Database      │◀───│  Analyzers     │
                       └──────────────────┘    └─────────────────┘
                                │                        │
                                ▼                        ▼
                       ┌──────────────────┐    ┌─────────────────┐
                       │ Health Checks   │    │   Stories      │
                       └──────────────────┘    └─────────────────┘
```

## Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/yourorg/incident-teller.git
cd incident-teller

# Install dependencies
go mod tidy

# Build the binary
go build -o incident-teller ./cmd/incident-teller

# Run with default config
./incident-teller
```

### Configuration

Create a `config.yaml` file:

```yaml
server:
  host: "0.0.0.0"
  port: 8080

netdata:
  base_url: "http://localhost:19999"
  poll_interval: "10s"

ai:
  enabled: true
  model_type: "local"
  confidence_threshold: 0.7

database:
  type: "sqlite"
  sqlite_path: "./incident_teller.db"

observability:
  log_level: "info"
  enable_metrics: true
  metrics_port: 9090
```

### Running with Docker

```bash
# Build Docker image
docker build -t incident-teller .

# Run with config
docker run -p 8080:8080 -p 9090:9090 \
  -v $(pwd)/config.yaml:/app/config.yaml \
  incident-teller -config /app/config.yaml
```

## Configuration Options

### Server Configuration
- **host**: Server bind address (default: 0.0.0.0)
- **port**: HTTP server port (default: 8080)
- **read_timeout**: Request read timeout (default: 30s)
- **write_timeout**: Response write timeout (default: 30s)

### Netdata Configuration  
- **base_url**: Netdata API URL (default: http://localhost:19999)
- **poll_interval**: Alert polling frequency (default: 10s)
- **timeout**: API request timeout (default: 30s)
- **retry_count**: Number of retries on failure (default: 3)

### AI Configuration
- **enabled**: Enable AI/ML features (default: true)
- **model_type**: AI model type - "local" or "external" (default: local)
- **confidence_threshold**: Minimum confidence for predictions (0.0-1.0, default: 0.7)
- **prediction_timeout**: Max time for AI predictions (default: 10s)

### Database Configuration
- **type**: Database type - "sqlite", "postgres", "mysql", or "memory"
- **sqlite_path**: SQLite database file path (default: ./incident_teller.db)
- **host**: Database host for SQL databases (default: localhost)
- **port**: Database port (default varies by type)
- **database**: Database name (default: incident_teller)

### Observability Configuration
- **log_level**: Logging level - debug, info, warn, error (default: info)
- **log_format**: Log format - json or text (default: json)
- **enable_metrics**: Enable metrics collection (default: true)
- **metrics_port**: Prometheus metrics port (default: 9090)

## API Endpoints

### Health Check
```
GET /health
```
Returns overall system health with component status.

### Metrics (if enabled)
```
GET /metrics
```
Prometheus-compatible metrics endpoint.

## Environment Variables

All configuration options can be overridden with environment variables:

```bash
# Server
export SERVER_HOST=0.0.0.0
export SERVER_PORT=8080

# Netdata  
export NETDATA_BASE_URL=http://localhost:19999
export NETDATA_POLL_INTERVAL=10s

# AI
export AI_ENABLED=true
export AI_CONFIDENCE_THRESHOLD=0.7

# Database
export DB_TYPE=sqlite
export DB_SQLITE_PATH=./incident_teller.db

# Observability
export OBSERVABILITY_LOG_LEVEL=info
export OBSERVABILITY_ENABLE_METRICS=true
```

## Production Deployment

### Kubernetes

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: incident-teller
spec:
  replicas: 2
  selector:
    matchLabels:
      app: incident-teller
  template:
    metadata:
      labels:
        app: incident-teller
    spec:
      containers:
      - name: incident-teller
        image: incident-teller:latest
        ports:
        - containerPort: 8080
        - containerPort: 9090
        env:
        - name: DB_TYPE
          value: "postgres"
        - name: DB_HOST
          value: "postgres-service"
        - name: AI_ENABLED
          value: "true"
        - name: OBSERVABILITY_LOG_LEVEL
          value: "info"
```

### Docker Compose

```yaml
version: '3.8'
services:
  incident-teller:
    build: .
    ports:
      - "8080:8080"
      - "9090:9090"
    environment:
      - DB_TYPE=postgres
      - DB_HOST=postgres
      - DB_USER=incident_teller
      - DB_PASSWORD=secure_password
      - NETDATA_BASE_URL=http://netdata:19999
      - AI_ENABLED=true
    depends_on:
      - postgres

  postgres:
    image: postgres:15
    environment:
      - POSTGRES_DB=incident_teller
      - POSTGRES_USER=incident_teller
      - POSTGRES_PASSWORD=secure_password
    volumes:
      - postgres_data:/var/lib/postgresql/data

volumes:
  postgres_data:
```

## Monitoring and Observability

### Metrics
- `incident_teller_uptime_seconds`: Application uptime
- `incident_teller_build_info`: Build information
- `alerts_received_total`: Number of alerts processed
- `ai_predictions_total`: Number of AI predictions made
- `incidents_analyzed_total`: Number of incidents analyzed

### Health Checks
- **Database**: Connection health and query performance
- **Netdata**: API reachability and response time  
- **Memory**: Memory usage and health status
- **Overall**: Aggregated system health status

### Structured Logging
All logs are structured with fields for:
- **timestamp**: RFC3339 timestamp
- **level**: Log level (debug, info, warn, error, fatal)
- **service**: Service name
- **request_id**: Request correlation ID (if applicable)
- **trace_id**: Distributed trace ID (if applicable)
- **error**: Error details (for error logs)

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run integration tests
go test -tags=integration ./...
```

### Building

```bash
# Build for current platform
go build -o incident-teller ./cmd/incident-teller

# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build -o incident-teller-linux ./cmd/incident-teller
GOOS=darwin GOARCH=amd64 go build -o incident-teller-darwin ./cmd/incident-teller
GOOS=windows GOARCH=amd64 go build -o incident-teller.exe ./cmd/incident-teller
```

### Development Mode

```bash
# Run with development config
export DB_TYPE=memory
export OBSERVABILITY_LOG_LEVEL=debug
export AI_ENABLED=true

./incident-teller
```

## Security Considerations

### Database Security
- Use connection pooling and proper connection limits
- Enable SSL/TLS for database connections in production
- Use database-specific user accounts with minimal privileges
- Regularly rotate database credentials

### API Security
- All health endpoints are read-only
- No sensitive data is exposed in metrics
- Logs may contain incident data - ensure proper log rotation
- Configure proper network access controls

### Configuration Security
- Store sensitive configuration in environment variables or secret managers
- Use read-only configuration files
- Audit configuration changes
- Validate all configuration values on startup

## Troubleshooting

### Common Issues

**AI Model Not Loading**
- Check `AI_ENABLED=true` in environment or config
- Verify model path permissions
- Check logs for model loading errors

**Database Connection Issues**
- Verify database is running and accessible
- Check connection string format and credentials
- Ensure database schema is initialized

**High Memory Usage**
- Reduce `DB_MAX_CONNECTIONS` if using SQL database
- Enable alert deduplication with `INCIDENT_ENABLE_ALERT_DEDUP=true`
- Adjust correlation window size

**Missing Alerts**
- Verify Netdata URL is correct and accessible
- Check Netdata alarm log configuration
- Verify `NETDATA_HOSTNAME` matches alert hostnames

### Debug Mode

Enable debug logging for detailed troubleshooting:

```bash
export OBSERVABILITY_LOG_LEVEL=debug
./incident-teller
```

This will show:
- Database query details
- AI prediction details
- Alert processing steps
- Configuration values

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit changes: `git commit -m 'Add amazing feature'`
4. Push to branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

### Development Guidelines
- Follow Go best practices and idioms
- Add tests for new functionality
- Update documentation for API changes
- Ensure all tests pass before submitting

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Documentation**: See this README and inline code comments
- **Issues**: Report bugs and feature requests on GitHub Issues
- **Discussions**: Use GitHub Discussions for questions and ideas

---

**IncidentTeller** - Transform alerts into intelligence. Built for SREs, by SREs.