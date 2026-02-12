# Notification Service

An event-driven notification system built with Go that accepts notification requests via HTTP, enqueues them to RabbitMQ, and delivers them to external webhook endpoints with rate limiting and retry logic.

## Features

- **Event-driven**: RabbitMQ topic exchange with channel-based queues (SMS, email, push) and priority support
- **Status tracking**: Full lifecycle (PENDING → QUEUED → SENT / FAILED / CANCELLED)
- **Retry logic**: Exponential backoff (up to 5 attempts) and DLQ for failed messages
- **Rate limiting**: Redis-based per-channel limit (e.g. 100 msg/sec)
- **Idempotency**: Redis + DB hybrid to prevent duplicate requests
- **Clean Architecture**: Domain, application (use cases), infrastructure, HTTP layers
- **Observability**: Health checks (DB, Redis), metrics (notification stats, queue depths)
- **Production ready**: Docker Compose with health checks, API + worker binaries

## Architecture

The project follows **Clean Architecture**:

- **Domain**: Entities, value objects, business rules
- **Application**: Use cases (create, cancel, get, list, process); ports (interfaces)
- **Infrastructure**: PostgreSQL (GORM), Redis, RabbitMQ, webhook client
- **HTTP**: Echo handlers, DTOs, middleware (correlation ID)

## Prerequisites

- **Go 1.22+**
- **Docker and Docker Compose**

## Quick Start

### 1. Clone and setup

```bash
git clone https://github.com/semih-yildiz/notification-service.git
cd notification-service
```

### 2. Run with Docker Compose

```bash
docker compose up --build
```

The API will be available at **http://localhost:8080**. The worker consumes from RabbitMQ and delivers to the configured webhook URL.

## Scalability: API and Worker Separation

The system is split into two separate processes (**API** and **Worker**) for scalability and resilience:

| Advantage | Description |
|-----------|-------------|
| **Independent scaling** | Scale API replicas for traffic and Worker replicas for throughput; each can be sized and scaled independently. |
| **Fault isolation** | A crash or overload in the worker does not take down the API; clients can still create and query notifications. |
| **Resource separation** | API focuses on HTTP and DB; Worker focuses on consumption and delivery. CPU/memory can be tuned per role. |
| **Deployment flexibility** | Deploy or roll back API and Worker on different schedules; run Workers only in regions that need delivery. |
| **Backpressure handling** | RabbitMQ queues absorb spikes; Workers drain at their own pace without blocking API responses. |

In production you can run multiple Worker instances (e.g. `docker compose up --scale worker=3`) to process queues in parallel; each message is consumed by a single worker.

## Configuration

Configuration is via **environment variables**. For local development you can use a `.env` file (e.g. copy `.env.example` to `.env.local`). Docker Compose sets variables for the stack.

### Required

| Variable      | Description                    | Example |
|---------------|--------------------------------|---------|
| `APP_PORT`    | HTTP server port               | `8080`  |
| `DB_DSN`      | PostgreSQL DSN                 | `postgres://postgres:postgres@postgres:5432/notifications?sslmode=disable` |
| `WEBHOOK_URL` | Delivery target for the worker | `https://webhook.site/your-unique-id` |

### Optional (with defaults)

| Variable                  | Description              | Default |
|---------------------------|--------------------------|---------|
| `APP_ENV`                 | Environment (local/test/prod) | `local` |
| `REDIS_ADDR`              | Redis address            | `localhost:6379` |
| `RABBITMQ_URL`            | RabbitMQ AMQP URL        | `amqp://guest:guest@localhost:5672/` |
| `RABBITMQ_MANAGEMENT_*`   | Management API (for /metrics) | `http://localhost:15672`, `guest`, `guest` |

### Docker

```yaml
# docker-compose.yml sets:
# - DB_DSN → postgres:5432, Redis → redis:6379, RabbitMQ → rabbitmq:5672
# - WEBHOOK_URL for worker delivery
```

### Local (without Docker)

1. Copy `.env.example` to `.env.local`.
2. Set `DB_DSN` to your local Postgres (e.g. `localhost:5432`).
3. Set `REDIS_ADDR` and `RABBITMQ_URL` to local instances.
4. Run API: `go run ./cmd/api` and Worker: `go run ./cmd/worker`.

## API

- **Base URL**: `http://localhost:8080`
- **Health**: `GET /health`
- **Metrics**: `GET /metrics` (notification counts, queue depths)

### Notifications & Batches

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST   | `/notifications` | Create single notification |
| POST   | `/notifications/batches` | Create batch (1–1000 items) |
| GET    | `/notifications/:id` | Get notification by ID |
| GET    | `/notifications` | List with filters (status, channel, batch_id, from, to, limit, offset) |
| POST   | `/notifications/:id/cancel` | Cancel pending notification |
| GET    | `/batches/:id/notifications` | Get batch and its notifications |
| POST   | `/batches/:id/cancel` | Cancel all pending in batch |

### Example: Create notification

```bash
curl -X POST http://localhost:8080/notifications \
  -H "Content-Type: application/json" \
  -d '{
    "recipient": "+905551234567",
    "channel": "sms",
    "content": "Your code is 123456",
    "priority": "high"
  }'
```

### Example: List notifications

```bash
curl "http://localhost:8080/notifications?limit=10&offset=0&status=pending"
```

## Project structure

```
notification-service/
├── cmd/
│   ├── api/          # HTTP API entrypoint
│   └── worker/       # RabbitMQ consumer entrypoint
├── internal/
│   ├── domain/notification/    # Entities, status, channel, rules
│   ├── application/notification/   # Use cases (create, cancel, get, list, process)
│   │   ├── command/  # create, cancel, process
│   │   ├── query/    # get, list
│   │   └── port/     # Repository, Publisher, Logger, etc.
│   ├── http/         # Echo routes, handlers, DTOs, middleware
│   └── infrastructure/
│       ├── config/   # Env-based config loader
│       ├── persistence/postgres/  # GORM, repos, migrations
│       ├── cache/redis/           # Client, idempotency, rate limiter
│       ├── messaging/rabbitmq/    # Publisher, consumer, topology
│       └── provider/webhook/     # Delivery client
├── docs/             # API and design docs
├── docker-compose.yml
├── Dockerfile        # Multi-stage build: api + worker
├── .env.example
└── go.mod / go.sum
```

## Database schema

- **batches**: One row per batch; notifications can optionally reference a batch.
- **notifications**: One row per notification; status lifecycle (pending → queued → sent/failed/cancelled).
- **delivery_attempts**: One row per delivery attempt (worker retries); linked to notifications.

### Database design 

```mermaid
erDiagram
    batches ||--o{ notifications : "batch_id"
    notifications ||--o{ delivery_attempts : "notification_id"

    batches {
        string id PK
        string idempotency_key UK
        datetime created_at
    }

    notifications {
        string id PK
        string batch_id FK
        string recipient
        string channel
        string content
        string priority
        string status
        string idempotency_key UK
        datetime created_at
        datetime updated_at
        datetime sent_at
        string failure_reason
    }

    delivery_attempts {
        string id PK
        string notification_id FK
        int attempt_number
        boolean success
        int status_code
        string response_body
        string error_message
        datetime created_at
    }
```


## Development

```bash
# Run API only (needs DB, Redis, RabbitMQ)
go run ./cmd/api

# Run worker only
go run ./cmd/worker

# Build both
go build -o api ./cmd/api && go build -o worker ./cmd/worker

# Tests
go test ./...
```
### System overview

```mermaid
---
config:
  theme: dark
  look: classic
---
flowchart LR
 subgraph subGraph0["Client Layer"]
        A["Mobile App"]
        B["Web App"]
        C["Backend Service"]
  end
 subgraph subGraph1["API Layer"]
        D["HTTP API<br>Port: 8080"]
        E["Middleware"]
        F["Router"]
  end
 subgraph subGraph2["Application Layer"]
        G["Create Notification"]
        H["Process UseCase"]
        I["Cancel UseCase"]
        J["Get UseCase"]
        K["List UseCase"]
        HC["Health Check<br>GET /health"]
        MET["Metrics<br>GET /metrics"]
  end
 subgraph subGraph3["Infrastructure - MQ"]
        L["RabbitMQ"]
        M["Queue: SMS"]
        N["Queue: Email"]
        O["Queue: Push"]
  end
 subgraph subGraph4["Worker Layer"]
        P["Consumer 1: SMS"]
        Q["Consumer 2: Email"]
        R["Consumer 3: Push"]
  end
 subgraph subGraph5["Infrastructure - Data"]
        S[("PostgreSQL<br>Notifications")]
        T[("Redis<br>Cache &amp; Rate Limit")]
  end
 subgraph subGraph6["External Services"]
        U["SMS Provider<br>Webhook"]
        V["Email Provider<br>Webhook"]
        W["Push Provider<br>Webhook"]
  end
    A --> D
    B --> D
    C --> D
    D --> E
    E --> F
    F --> G & I & J & K & HC & MET
    G --> S & T & L
    L --> M & N & O
    M --> P
    N --> Q
    O --> R
    P --> H
    Q --> H
    R --> H
    H --> S & T & U & V & W
    J --> S
    K --> S
    I --> S
    HC --> S & T
    MET --> S & L

    style D fill:#38BDF8,stroke:#E2E8F0,stroke-width:1px
    style HC fill:#818CF8,stroke:#000000,stroke-width:1px,color:#424242
    style MET fill:#818CF8,stroke:#000000,stroke-width:1px,color:#424242
    style L fill:#757575,stroke:#000000,stroke-width:1px,color:#000000
    style P fill:#4ADE80,stroke:#000000,stroke-width:1px,color:#000000
    style Q fill:#4ADE80,stroke:#000000,stroke-width:1px,color:#000000
    style R fill:#4ADE80,stroke:#000000,stroke-width:1px,color:#000000
    style S fill:#A78BFA,stroke:#E2E8F0,stroke-width:1px,color:#000000
    style T fill:#F87171,stroke:#E2E8F0,stroke-width:1px
    style subGraph4 fill:#424242,stroke:#FFD600
    style subGraph0 fill:#424242,stroke:#424242
    style subGraph1 stroke:#FFD600

```
