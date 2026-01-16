# IncidentTeller Architecture

## 1. High-Level Architecture

IncidentTeller follows a **Clean Architecture (Hexagonal)** approach to decouple the business logic from the data source (Netdata) and the presentation layer.

### Components:
1.  **Ingestion Worker (Poller)**:
    - Periodically polls Netdata's `/api/v1/alarm_log`.
    - Tracks the `last_unique_id` to ensure exactly-once processing of alert logs.
    - Normalizes Netdata response into a generic `Alert` event.

2.  **Event Processor (Domain Service)**:
    - **Correlation**: Groups alerts based on `chart_id`, `family`, and `server`.
    - **State Management**: Tracks the lifecycle of an incident (e.g., `WARNING` -> `CRITICAL` -> `CLEAR`).
    - **Debouncing**: Handles flapping alerts to avoid spammy timelines.

3.  **Timeline Generator**:
    - Converts a sequence of Alert events into human-readable sentences.
    - Example: _"High CPU usage detected (95%) at 10:00 AM, escalated to Critical at 10:05 AM, resolved at 10:15 AM."_

4.  **Storage (Repository)**:
    - Stores raw events and active incident states. (Interfaces defined in Core, implemented as In-Memory or SQL).

## 2. Data Flow
`Netdata API` -> `Poller` -> `Event Channel` -> `Processor` -> `Repository` -> `Timeline Service` -> `Output (JSON/CLI)`

---

## 3. Go Folder Structure

```text
IncidentTeller/
├── cmd/
│   └── incident-teller/
│       └── main.go           # Entry point, dependency injection, wiring
├── internal/
│   ├── domain/               # Core business logic (Pure Go, no external deps)
│   │   ├── models.go         # Alert, Incident, Timeline structs
│   │   └── service.go        # Timeline generation logic
│   ├── ports/                # Interfaces (Secondary Ports)
│   │   ├── repository.go     # Writer/Reader interfaces
│   │   └── alert_source.go   # Interface for fetching alerts
│   └── adapters/             # Implementation of ports
│       ├── netdata/          # Netdata API client
│       │   ├── client.go
│       │   └── mapper.go     # Maps JSON -> domain.Alert
│       └── memory/           # In-memory storage (for MVP)
│           └── store.go
├── pkg/
│   └── utils/                # Shared utilities (Logger, Time helper)
└── config/
    └── config.go             # Configuration loading (Netdata URL, Polling interval)
```
