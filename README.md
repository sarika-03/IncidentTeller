# IncidentTeller - AI-Powered SRE Incident Analysis Platform

**IncidentTeller** is a comprehensive incident analysis platform that transforms raw alert data into actionable intelligence for SRE teams. It combines a Go backend for intelligent analysis with a modern Next.js frontend for visualization and real-time monitoring.

## 📋 What This Project Does

IncidentTeller helps on-call engineers:
- **Correlate Alerts**: Groups related alerts into incidents automatically
- **Root Cause Analysis**: Uses AI/ML to predict the primary root cause of incidents
- **Blast Radius Detection**: Identifies affected services and impact scope
- **Timeline Visualization**: Shows incident progression with clear causality
- **Fix Recommendations**: Provides actionable remediation steps based on incident type
- **Health Monitoring**: Tracks system health with real-time status checks

The system continuously polls monitoring data (currently supports Netdata), correlates alerts, stores them persistently, and provides insights through an intuitive web interface.

## 🏗️ Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                    Frontend (Next.js)                        │
│  http://localhost:3000                                       │
│  - Dashboard with incident summary                           │
│  - Incident list with pagination                             │
│  - Incident detail page with AI analysis                     │
│  - Health check page                                         │
│  - Timeline visualization                                    │
└──────────────────┬───────────────────────────────────────────┘
                   │ HTTP/REST API
                   ▼
┌──────────────────────────────────────────────────────────────┐
│                  Backend (Go) API                            │
│  http://localhost:8080/api                                   │
│                                                              │
│  ┌────────────────┐  ┌──────────────┐  ┌──────────────────┐ │
│  │ Alert Poller   │  │ AI Models    │  │ Incident Builder│ │
│  │ (Netdata)      │  │ (Local ML)   │  │ (Correlation)  │ │
│  └────────┬───────┘  └──────┬───────┘  └────────┬─────────┘ │
│           │                 │                     │           │
│           └─────────────────┼─────────────────────┘           │
│                             ▼                                 │
│                  ┌──────────────────────┐                     │
│                  │  Repository/Storage  │                     │
│                  │  (SQLite/Memory)     │                     │
│                  └──────────────────────┘                     │
│                                                              │
│  Routes:                                                     │
│  - GET  /api/health                (System health)           │
│  - GET  /api/incidents             (List all incidents)      │
│  - GET  /api/incidents/{id}        (Incident detail)         │
│  - GET  /api/incidents/summary     (Stats & summary)         │
│  - GET  /api/timeline/{id}         (Event timeline)          │
└──────────────────────────────────────────────────────────────┘
                   │
                   ▼
┌──────────────────────────────────────────────────────────────┐
│                   Data Sources                               │
│                                                              │
│  ┌──────────────────┐           ┌──────────────────────┐    │
│  │  Netdata Agent   │           │ SQLite Database      │    │
│  │ (localhost:19999)│           │ (incident_teller.db) │    │
│  └──────────────────┘           └──────────────────────┘    │
└──────────────────────────────────────────────────────────────┘
```

## 🚀 Quick Start

### Prerequisites
- **Go 1.19+** (for backend)
- **Node.js 16+** and **npm** (for frontend)
- **Netdata** (for alert data source) - optional for demo mode

### Backend Setup

```bash
# 1. Build the backend
cd /home/sarika/IncidentTeller
go build -o incident-teller main.go

# 2. Configure (uses config.yaml by default)
# Default config uses SQLite and in-memory repository

# 3. Run the backend
./incident-teller

# Backend will start on http://localhost:8080
# API available at http://localhost:8080/api
```

The backend will:
- Listen on port 8080
- Poll Netdata for alerts every 10 seconds (configurable)
- Store incidents in SQLite or memory
- Expose REST API endpoints
- Run health checks automatically

### Frontend Setup

```bash
# 1. Navigate to UI folder
cd /home/sarika/IncidentTeller/ui

# 2. Install dependencies
npm install

# 3. Run development server
npm run dev

# Frontend will start on http://localhost:3000
```

The frontend will:
- Start on port 3000
- Auto-reload on code changes (hot module reloading)
- Connect to backend API at http://localhost:8080/api
- Display incidents, analytics, and health status in real-time

### Verify Installation

Open your browser and navigate to:
- **Dashboard**: http://localhost:3000
- **Incidents**: http://localhost:3000/incidents
- **Health Check**: http://localhost:3000/health
- **Health API**: http://localhost:8080/api/health

## 📁 Project Structure

```
IncidentTeller/
├── main.go                          # Backend entry point
├── config.yaml                      # Configuration file
├── cmd/
│   ├── demo-generator/             # Demo data generator
│   └── incident-teller/            # Main application
├── internal/
│   ├── adapters/
│   │   ├── netdata/               # Netdata API client
│   │   ├── openai/                # OpenAI integration
│   │   └── repository/            # In-memory storage
│   ├── ai/                        # AI/ML models for analysis
│   ├── api/                       # HTTP handlers & routing
│   ├── config/                    # Configuration management
│   ├── database/                  # Database/repository layer
│   ├── domain/                    # Core domain models
│   ├── observability/             # Logging, metrics, health checks
│   ├── ports/                     # Interface definitions
│   └── services/                  # Business logic services
│       ├── analyzer.go            # Alert analysis
│       ├── blast_radius_analyzer  # Impact analysis
│       ├── incident_builder.go    # Incident correlation
│       └── timeline_builder.go    # Timeline generation
├── ui/                            # Frontend (Next.js)
│   ├── src/
│   │   ├── app/                  # Pages and layouts
│   │   │   ├── page.tsx          # Dashboard
│   │   │   ├── incidents/        # Incidents list
│   │   │   ├── health/           # Health page
│   │   │   └── timeline/         # Timeline page
│   │   ├── components/           # React components
│   │   ├── lib/                  # Utilities (API client)
│   │   └── types/                # TypeScript types
│   └── package.json              # Dependencies
└── examples/                      # Demo scripts
```

## 🔧 Configuration

### Backend Configuration (config.yaml)

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
  type: "sqlite"              # or "memory"
  sqlite_path: "./incident_teller.db"

observability:
  log_level: "info"
  enable_metrics: true
```

### Environment Variables

```bash
# Backend
export SERVER_PORT=8080
export DB_TYPE=sqlite
export AI_ENABLED=true
export NETDATA_BASE_URL=http://localhost:19999

# Frontend  
export NEXT_PUBLIC_API_URL=http://localhost:8080/api
```

## 📊 Core Features

### 1. Alert Correlation
- Automatically groups related alerts into incidents
- Configurable correlation window (default: 15 minutes)
- Deduplication of duplicate alerts

### 2. Root Cause Analysis
- AI-powered prediction of primary root cause
- Confidence scores (0-100%)
- Pattern recognition (burst, cascade, gradual, spike)
- Alternative causes with reasoning

### 3. Blast Radius Analysis
- Estimates impact scope and affected services
- Cascade probability calculation
- Business impact assessment
- Duration prediction

### 4. Real-Time Updates
- Server-Sent Events (SSE) for live updates
- Polling fallback (30 seconds)
- Real-time incident status changes

### 5. Health Monitoring
- Database connection health
- Netdata API reachability
- Memory usage monitoring
- System-level health aggregation

## 📈 API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/health` | System health status |
| GET | `/api/incidents` | List all incidents (paginated) |
| GET | `/api/incidents/{id}` | Incident details with AI analysis |
| GET | `/api/incidents/summary` | Summary statistics |
| GET | `/api/timeline/{id}` | Incident timeline events |
| POST | `/api/subscribe` | WebSocket/SSE for real-time updates |

## 🎯 How It Works - Step by Step

### 1. Alert Detection
- Backend polls Netdata API every 10 seconds
- Fetches all active alerts from monitoring system
- Stores alerts in repository

### 2. Incident Correlation
- Groups alerts by time window (15-minute window)
- Links related alerts to same incident
- Updates incident status based on alert patterns

### 3. AI Analysis
- **Root Cause**: ML model analyzes alert patterns
- **Blast Radius**: Predicts service impact
- **Fix Recommendations**: Suggests remediation based on cause

### 4. Data Persistence
- SQLite stores incidents and alerts
- In-memory mode for development/testing
- Automatic schema initialization

### 5. Frontend Display
- Fetches incident data from backend
- Displays real-time updates via SSE
- Shows AI analysis and recommendations
- Visualizes incident timeline

## 🔍 Monitoring and Debugging

### Check Backend Health
```bash
curl http://localhost:8080/api/health | jq .
```

### View Logs
```bash
tail -f incident-teller.log
```

### Access Database
```bash
sqlite3 incident_teller.db

# List all incidents
SELECT id, title, status, started_at FROM incidents;

# Count incidents by status
SELECT status, COUNT(*) FROM incidents GROUP BY status;
```

### Frontend Console
- Open browser DevTools (F12)
- Check Console tab for API errors
- Check Network tab to see API requests

## 📚 Key Services

### AlertGrouper
Groups multiple alerts into logical incident groups for correlation analysis.

### IncidentBuilder
Correlates related alerts into incidents using time-window based correlation.

### Analyzer
Performs statistical analysis on incident patterns and alert characteristics.

### BlastRadiusAnalyzer
Estimates the scope and impact of incidents on system components.

### TimelineBuilder
Constructs chronological event timelines for incidents with causality information.

## 🛠️ Development

### Run Backend in Debug Mode
```bash
export OBSERVABILITY_LOG_LEVEL=debug
./incident-teller
```

### Run Frontend with HMR
```bash
cd ui
npm run dev
```

### Generate Demo Data
```bash
cd /home/sarika/IncidentTeller/cmd/demo-generator
go run main.go
```

## ⚙️ Performance Considerations

- **Alert Processing**: Optimized for 1000+ alerts per minute
- **Correlation Window**: Default 15 minutes, adjustable
- **Database**: SQLite suitable for up to 100K incidents; use PostgreSQL for larger volumes
- **Memory Usage**: In-memory mode uses ~50MB for 10K incidents
- **API Response Time**: Sub-100ms for typical queries

## 🔐 Security

- Read-only health endpoints
- No sensitive data in logs
- SQLite file permissions (600)
- Environment-based secrets
- Input validation on all API endpoints

## 🐛 Troubleshooting

### Backend won't start
```bash
# Check port 8080 is free
lsof -i :8080

# Check config.yaml syntax
cat config.yaml

# Enable debug logging
export OBSERVABILITY_LOG_LEVEL=debug
```

### Frontend can't connect to backend
```bash
# Verify backend is running
curl http://localhost:8080/api/health

# Check CORS headers
curl -v http://localhost:8080/api/incidents
```

### No incidents appearing
- Check Netdata is running: `curl http://localhost:19999`
- Check alert configuration in Netdata
- Verify config.yaml `netdata.base_url` is correct

## 📞 Support

- Check logs: `tail -f incident-teller.log`
- View configuration: `cat config.yaml`
- Test API: `curl http://localhost:8080/api/health | jq .`
- Frontend errors: Open DevTools (F12) in browser

---

**IncidentTeller** - AI-powered incident analysis for modern SRE teams. Built with Go backend and Next.js frontend.