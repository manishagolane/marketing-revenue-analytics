Marketing & Revenue Analytics System

1. Problem Statement

Design a scalable Marketing & Revenue Analytics backend system that:

Tracks campaign performance (impressions, clicks, conversions)
Processes high-volume user events in real-time
Provides aggregated analytics (daily, weekly, monthly)
Supports campaign management (CRUD + RBAC)
Exposes public + authenticated APIs
Handles async event ingestion reliably

2. Core System Goals
Functional Requirements
Campaign CRUD
Event tracking API (impression, click, conversion)
Funnel analytics
Time spent analytics
Click path tracking
Daily/weekly/monthly aggregation
Public campaign insights
Non-Functional Requirements
High throughput event ingestion
Low latency API responses (<200ms target)
Event durability (no loss)
Horizontally scalable consumers
Fault tolerant processing
Cache optimization for analytics

```text
3. High Level Architecture

                ┌──────────────────────┐
                │   Client Apps        │
                └─────────┬────────────┘
                          │
                          ▼
                ┌──────────────────────┐
                │   API Gateway (Gin)  │
                └─────────┬────────────┘
                          │
        ┌────────────────────────────────────┐
        │            Backend System          │
        └────────────────────────────────────┘
          │            │             │
          ▼            ▼             ▼
 ┌─────────────┐ ┌────────────┐ ┌──────────────┐
 │ Auth Service│ │ Campaign   │ │ Analytics    │
 │             │ │ Service    │ │ Service      │
 └──────┬──────┘ └────┬───────┘ └────┬─────────┘
        │             │              │
        ▼             ▼              ▼
     PostgreSQL   PostgreSQL   Aggregation DB
        │
        ▼
 ┌──────────────────────────────┐
 │ Event Ingestion Pipeline     │
 └─────────────┬────────────────┘
               ▼
        ┌──────────────┐
        │ AWS SQS Queue│  (Event Buffer)
        └──────┬───────┘
               ▼
     ┌───────────────────────┐
     │ Event Consumer Worker │
     └─────────┬─────────────┘
               ▼
     ┌───────────────────────┐
     │ Event Logs DB         │
     └───────────────────────┘
```

4. Low Level Architecture 
📁 Project Structure
cmd/app/                 → main entrypoint
internal/
  handlers/              → HTTP layer
  dto/                   → request/response contracts
  service/               → business logic (optional expansion)
  clients/               → external systems (SQS, cache)
  consumer/              → async event worker
  server/                → router setup
db/                      → SQL schema + migrations
models/                  → sqlc generated models
utils/                   → helpers
cache/                   → Redis abstraction
config/                  → env config
logger/                  → structured logging (zap)

5. 🔄 Request Flow Design
Example: Event Tracking Flow
Client → /events/track
        ↓
Handler validates request
        ↓
Check campaign status
        ↓
Transform → Event DTO
        ↓
Push to SQS (async)
        ↓
Return 202 Accepted


Consumer Flow
SQS Message
    ↓
Consumer Worker
    ↓
Validate event
    ↓
Insert into event_logs
    ↓
Update campaign_daily_metrics
    ↓
Log success/failure


6. Database Design (Schema Decisions)
👤 users table
Stores platform users (admin, marketer, analyst)
Role-based access control (RBAC)


campaigns table
- id (ULID)
- created_by → FK users
- status (draft/active/paused/completed)
- budget, spend, revenue

Why:
Supports lifecycle tracking
Enables analytics grouping
Soft delete supported


📊 event_logs table
- campaign_id
- event_type (impression/click/conversion)
- metadata (JSONB)
- session tracking

Why JSONB?
Flexible event metadata (future-proof)
Avoid schema changes for new tracking attributes


📈 campaign_daily_metrics
Pre-aggregated table
Used for fast analytics queries
Why:
Avoid real-time aggregation on large event_logs
Enables fast dashboard queries


7. Event System Design
Why Event Driven Architecture?

Because:
Event volume is high (click/impression scale)
Sync DB writes would bottleneck API
We need decoupling between ingestion and analytics


8. Why AWS SQS (NOT Kafka)
Decision: SQS over Kafka
Reasons for SQS
1. Operational simplicity
Fully managed
No cluster maintenance
No partition tuning

2. Perfect for our use case
At-least-once delivery is enough
Order is NOT critical
Event stream is not replay-heavy

3. Cost efficiency
Pay-per-message model
No always-on brokers

4. Horizontal scaling ready
Multiple consumers can process independently
❌ Why NOT Kafka

Kafka is overkill because:
Requires cluster management
Partition design complexity
Replay requirement not needed here
Higher infra cost
📌 Future upgrade path

If system scales:
SQS → Kafka (event streaming + replay + real-time analytics)

9. 🧾 Event Processing Design
Event Payload
{
  "campaign_id": "c1",
  "event_type": "click",
  "source_url": "google.com",
  "session_id": "abc123",
  "step": "landing",
  "metadata": {}
}

Processing Strategy
Step 1: Validation
campaign exists
campaign is active
Step 2: Async ingestion
push to SQS
Step 3: Consumer processing
insert into event_logs
update campaign_daily_metrics


10. Architectural Principles Used
1. Hexagonal Architecture (Ports & Adapters)
Handlers → Service → Repository → DB

External systems:
SQS
Redis
PostgreSQL

All behind interfaces.

2. Dependency Inversion
handlers depend on interfaces
not concrete DB implementations

3. Separation of Concerns
Layer	    Responsibility
Handler	    HTTP + validation
Service	    business logic
Repository	DB access
Consumer	async processing

4. DTO Pattern
request validation isolated
response shaping independent of DB models

11. Caching Strategy
Used for:
analytics endpoints
daily metrics
campaign summary
Cache key example:
analytics:daily:{campaign}:{from}:{to}:{page}:{limit}
TTL:
5 minutes (balance freshness + performance)


12. Testing Strategy
Unit Tests
handler validation
service logic
funnel calculations
drop-off logic
Integration Tests
API → DB
SQS → consumer → DB
Mocking Strategy
interfaces for all stores
mock cache
mock queue

13. Scalability Design
Horizontal Scaling
Component	Scaling Strategy
API	stateless pods
Consumer	multiple workers
DB	read replicas
Cache	Redis cluster
Event Scaling
SQS supports unlimited throughput
consumer autoscaling based on queue depth

14. Security Design
JWT authentication (stateless)
Role-based access control
Public vs private campaign APIs
Input validation at handler level
SQL injection safe via sqlc


15. 📊 Observability
Logging
zap structured logs





## Table of Contents

- [Tech Stack](#tech-stack)
- [Prerequisites](#prerequisites)
- [Configuration](#configuration)
- [Database Setup](#database-setup)
- [Running the Server](#running-the-server)
- [Project Structure](#project-structure)
- [Architecture](#architecture)
- [API Reference](#api-reference)
- [Metrics & Formulas](#metrics--formulas)
- [Event Tracking Flow](#event-tracking-flow)
- [Caching Strategy](#caching-strategy)
- [Rate Limiting](#rate-limiting)
- [Role-Based Access Control](#role-based-access-control)
- [Think Pieces](#think-pieces)
- [Future Improvements](#future-improvements)
- [AI Tooling](#ai-tooling)

---



## Prerequisites

Make sure the following are installed and running before starting:

- Go 1.22+
- PostgreSQL 15+
- Redis 7+
- AWS account with an SQS Standard Queue created
- AWS credentials configured (`~/.aws/credentials` or environment variables)

---

## Configuration

Copy the example config and fill in your values:

```bash
```


```yaml



db:
  host: localhost
  port: 5432
  user: postgres
  password: yourpassword
  dbname: marketing_analytics
  sslmode: disable
  timeout: 10

redis:
  address: localhost:6379
  password: ""
  db: 0
  lockDB: 1
  lockExpiry: 30

authentication:
  jwt:
    user_secret: ""    # 64-byte EdDSA private key as hex string
                       # generate: openssl genpkey -algorithm ed25519 | xxd -p -c 256

aws:
  region: us-east-1

event:
  sqs: https://sqs.us-east-1.amazonaws.com/572338935572/market-analytics-project

retry:
  maxRetries: 3
  baseDelay: 2

log:
  fileName: logs/app.log
  maxSize: 100
  maxBackups: 3
  maxAge: 28
```

### Generating a JWT Secret

```bash
# Generates a 64-byte Ed25519 private key in hex format
openssl genpkey -algorithm ed25519 -outform DER | tail -c 32 | cat <(openssl genpkey -algorithm ed25519 -outform DER) | xxd -p -c 256
```

Or use the Go playground to generate with `crypto/ed25519.GenerateKey()` and hex-encode the private key bytes.

---

## Database Setup

### 1. Create the database

```bash
psql -U postgres -c "CREATE DATABASE marketing_analytics;"
```

### 2. Run migrations in order

```bash
psql -U postgres -d marketing_analytics -f db/migrations/001_create_users.sql
psql -U postgres -d marketing_analytics -f db/migrations/002_create_campaigns.sql
psql -U postgres -d marketing_analytics -f db/migrations/003_create_event_logs.sql
psql -U postgres -d marketing_analytics -f db/migrations/004_create_campaign_daily_metrics.sql
```

### Schema Overview

```sql
-- Users: auth + RBAC
users (id, name, email, password, phone, role, bio, picture, created_at, updated_at, deleted_at)

-- Campaigns: lifecycle management
campaigns (id, name, description, created_by, status, channel, budget, spend, revenue, is_public, starts_at, ends_at, ...)

-- Raw event log: append-only time-series
event_logs (id, campaign_id, event_type, source_url, ip_address, user_agent, session_id, step, metadata, occurred_at)

-- Pre-aggregated daily rollup: drives all analytics queries
campaign_daily_metrics (campaign_id, date, impressions, clicks, conversions)
```

---

## Running the Server

```bash
# Install dependencies
go mod download

# Run
go run main.go

# Build binary
go build -o marketing-analytics main.go
./marketing-analytics
```

The server starts on the port defined in `config.yaml` (default `:8010`).

Health check:

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

---

## Project Structure

```
marketing-revenue-analytics/
├── main.go                        # Entry point — wires all dependencies
├── config/                        # Viper config loader
├── logger/                        # Zap structured logger init
├── auth/                          # EdDSA JWT manager
├── cache/                         # Redis abstraction (rueidis + redsync)
├── db/                            # PostgreSQL pool init + migrations
├── models/                        # sqlc-generated type-safe query layer
├── utils/                         # Shared helpers (ULID, null types, context)
├── constants/                     # App-wide constants (roles, keys, timezone)
├── internal/
│   ├── server/
│   │   ├── server.go              # HTTP server + graceful shutdown
│   │   ├── routes.go              # Route definitions
│   │   └── middleware.go          # Auth, RBAC, rate limit, logging, recovery
│   ├── handlers/
│   │   ├── auth.go                # Register, login, logout, profile
│   │   ├── campaign.go            # Campaign CRUD + status management
│   │   ├── event.go               # Event tracking endpoint
│   │   ├── analytics.go           # Daily/weekly/monthly/funnel analytics
│   │   ├── errors.go              # Standardised error responses
│   │   └── validation.go          # Pretty validation error messages
│   ├── events/
│   │   └── payload.go             # EventPayload DTO
│   ├── consumer/
│   │   └── consumer.go            # SQS consumer — inserts logs + upserts metrics
│   └── clients/
│       ├── sqs.go                 # AWS SQS client (send + poll + delete)
│       └── clients.go             # Dependency container
└── README.md
```

---

## Architecture

### High-Level Overview

```
Client Apps
    │
    ▼
Gin HTTP Server (API Gateway)
    │
    ├── Auth Handler       → PostgreSQL (users)
    ├── Campaign Handler   → PostgreSQL (campaigns)
    ├── Analytics Handler  → Redis cache → PostgreSQL (campaign_daily_metrics)
    └── Event Handler      → AWS SQS ──────────────────────────┐
                                                               │
                                                    SQS Consumer Worker
                                                               │
                                                    ┌──────────┴──────────┐
                                                    │  Insert event_logs  │
                                                    │  Upsert daily_metrics│
                                                    └─────────────────────┘
```

### Architecture Explanation

The system is split into two separate workload paths to isolate write-heavy event ingestion from read-heavy analytics reporting.

**Write path (event ingestion):** When a tracking event arrives at `POST /events/track`, the handler validates the request and immediately publishes the payload to AWS SQS, returning `202 Accepted` in under 5ms. A background consumer polls SQS continuously, inserts one raw row into `event_logs`, then atomically increments the `campaign_daily_metrics` rollup table via an `ON CONFLICT DO UPDATE` upsert. This means analytics queries never touch the raw event table — they only read the pre-aggregated daily rollup, which stays fast regardless of event volume.

**Read path (analytics):** Daily, weekly, and monthly analytics are served from `campaign_daily_metrics` with computed CTR, CPC, ROI, and conversion rate calculated inline by Postgres. Responses are cached in Redis (5 min / 10 min / 30 min TTLs respectively) so repeated dashboard reads don't hit the database. Authentication uses EdDSA-signed JWTs for stateless identity with Redis as the session store — on logout the JWT signature is deleted from Redis, invalidating the token instantly without waiting for expiry.

---

## API Reference

### Base URL

```
http://localhost:8010/api/v1
```

### Authentication

All protected endpoints require a JWT token in the `Authorization` header:

```
Authorization: <token>
```

---

### Auth

#### Register
```
POST /auth/register
```
All new users register as `marketer` by default. Role promotion is admin-only.

Request:
```json
{
  "name": "Jane Doe",
  "email": "jane@example.com",
  "password": "securepassword",
  "phone": "9876543210"
}
```

Response `201`:
```json
{
  "status": "success",
  "message": "registered successfully",
  "data": {
    "user": { "id": "...", "name": "Jane Doe", "role": "marketer", ... },
    "token": "<jwt>"
  }
}
```

---

#### Login
```
POST /auth/login
```

Request:
```json
{
  "email": "jane@example.com",
  "password": "securepassword"
}
```

Response `200`:
```json
{
  "status": "success",
  "data": { "token": "<jwt>", "user": { ... } }
}
```

---

#### Logout
```
POST /auth/logout
```
Invalidates the session in Redis immediately.

---

#### Get Profile
```
GET /auth/profile
```

---

#### Update Profile
```
PATCH /auth/profile
```

Request:
```json
{
  "name": "Jane Smith",
  "phone": "9876543210",
  "bio": "Senior marketer",
  "picture": "https://example.com/avatar.jpg"
}
```

---

#### Promote User Role (Admin only)
```
PATCH /auth/users/:id/role
```

Request:
```json
{
  "role": "analyst"
}
```

---

### Campaigns

#### Create Campaign
```
POST /campaigns
```
Roles: `admin`, `marketer`

Request:
```json
{
  "name": "Summer Sale 2025",
  "description": "Q3 revenue push",
  "channel": "email",
  "budget": 5000.00,
  "is_public": true,
  "starts_at": "2025-07-01T00:00:00Z",
  "ends_at": "2025-07-31T23:59:59Z"
}
```

---

#### List Campaigns
```
GET /campaigns?status=active&channel=email&from=2025-01-01T00:00:00Z&to=2025-12-31T23:59:59Z&page=1&limit=20
```

Filters: `status`, `channel`, `from`, `to`, `is_public`, `created_by` (admin only — marketers automatically scoped to own campaigns)

---

#### Search Campaigns
```
GET /campaigns/search?q=summer&page=1&limit=20
```
Full-text search across `name` and `description`.

---

#### Get Campaign
```
GET /campaigns/:id
```

---

#### Update Campaign
```
PATCH /campaigns/:id
```
Roles: owner or `admin`

---

#### Update Campaign Status
```
PATCH /campaigns/:id/status
```

Request:
```json
{
  "status": "active"
}
```

Valid statuses: `draft` → `active` → `paused` → `completed` → `archived`

---

#### Delete Campaign
```
DELETE /campaigns/:id
```
Soft delete. Roles: owner or `admin`

---

#### Public Campaign List (no auth)
```
GET /campaigns/public?page=1&limit=20
```

---

#### Public Campaign Preview (no auth)
```
GET /campaigns/:id/preview
```

---

### Event Tracking

#### Track Event (no auth — external systems)
```
POST /events/track
```

Request:
```json
{
  "campaign_id": "01J...",
  "event_type": "click",
  "source_url": "https://example.com/landing",
  "session_id": "sess_abc123",
  "step": "landing",
  "occurred_at": "2025-07-15T10:30:00Z",
  "metadata": {
    "browser": "chrome",
    "device": "mobile"
  }
}
```

Fields:
- `event_type`: `impression` | `click` | `conversion`
- `session_id`: groups events from the same user session (enables funnel tracking)
- `step`: funnel stage e.g. `ad`, `landing`, `signup`, `purchase`
- `occurred_at`: optional, defaults to server time

Response `202`:
```json
{
  "status": "success",
  "message": "event tracked",
  "data": { "event_id": "01J..." }
}
```

---

### Analytics

All analytics endpoints require auth. Roles: `admin`, `marketer`, `analyst`

Common query params: `campaign_id`, `channel`, `from` (YYYY-MM-DD), `to` (YYYY-MM-DD), `page`, `limit`

#### Daily Metrics
```
GET /analytics/daily?campaign_id=01J...&from=2025-07-01&to=2025-07-31
```

#### Weekly Metrics
```
GET /analytics/weekly?channel=email&from=2025-07-01&to=2025-07-31
```

#### Monthly Metrics
```
GET /analytics/monthly?campaign_id=01J...
```

Each response includes computed: `ctr`, `cpc`, `roi`, `conversion_rate`

---

#### Funnel Analytics
```
GET /analytics/campaigns/:id/funnel?page=1&limit=50
```

Response:
```json
{
  "status": "success",
  "data": {
    "funnel": [
      { "step": "ad",       "unique_sessions": 10000 },
      { "step": "landing",  "unique_sessions": 6500  },
      { "step": "signup",   "unique_sessions": 1800  },
      { "step": "purchase", "unique_sessions": 420   }
    ],
    "drop_offs": [
      { "from": "ad",      "to": "landing",  "drop_off%": "35.00" },
      { "from": "landing", "to": "signup",   "drop_off%": "72.31" },
      { "from": "signup",  "to": "purchase", "drop_off%": "76.67" }
    ],
    "sessions": [
      { "session_id": "abc", "session_start": "...", "session_end": "...", "duration_seconds": 142 }
    ]
  }
}
```

---

#### Public Campaign Summary (no auth)
```
GET /analytics/campaigns/:id/summary
```
Returns anonymized stats for public campaigns only.

---

### Error Responses

All errors follow a consistent shape:

```json
{
  "status": "failure",
  "message": "human readable error",
  "errorCode": 400
}
```

| Code | Meaning |
|---|---|
| 400 | Bad request / validation failed |
| 401 | Missing or invalid token |
| 403 | Insufficient role permissions |
| 404 | Resource not found |
| 409 | Conflict (e.g. duplicate email) |
| 429 | Rate limit exceeded |
| 500 | Internal server error |

---

## Metrics & Formulas

| Metric | Formula |
|---|---|
| CTR (Click-Through Rate) | `(clicks / impressions) × 100` |
| CPC (Cost Per Click) | `spend / clicks` |
| ROI (Return on Investment) | `((revenue - spend) / spend) × 100` |
| Conversion Rate | `(conversions / clicks) × 100` |

All computed inline by Postgres at query time using pre-aggregated daily metrics.

---

## Event Tracking Flow

```
POST /events/track
        │
        ▼
Validate request (campaign exists + active)
        │
        ▼
Build EventPayload (assign ULID, capture IP + user-agent)
        │
        ▼
Publish to AWS SQS
        │
        ▼
Return 202 Accepted  ← client gets response here, ~5ms

━━━━━━━━━━━━━━━━━━━━ async boundary ━━━━━━━━━━━━━━━━━━━━

SQS Consumer (background goroutine)
        │
        ▼
Deserialize EventPayload
        │
        ▼
INSERT into event_logs (idempotent — duplicate ULID = skip)
        │
        ▼
UPSERT campaign_daily_metrics (ON CONFLICT DO UPDATE, atomic increment)
        │
        ▼
Delete SQS message (success) or leave for retry (failure)
```

---

## Caching Strategy

| Endpoint           | Cache Key                                                               | TTL                |
|--------------------|-------------------------------------------------------------------------|--------------------|
| Daily analytics    | `analytics:daily:{campaign}:{channel}:{from}:{to}:{page}:{limit}`       | 5 min              |
| Weekly analytics   | `analytics:weekly:{campaign}:{channel}:{from}:{to}:{page}:{limit}`      | 10 min             |
| Monthly analytics  | `analytics:monthly:{campaign}:{channel}:{from}:{to}:{page}:{limit}`     | 30 min             |
| RBAC permissions   | `permission:{role}:{object}:{action}`                                   | 10 min             |
| JWT sessions       | `jwt_sig:{signature}`                                                   | 24 hr (token TTL)  |

---

## Rate Limiting

Implemented via Redis sliding counter (per 60-second window).

| Endpoint type            | Identifier | Limit      |
|--------------------------|------------|------------|
| Public / unauthenticated | IP address | 30 req/min |
| Authenticated            | User ID    | 60 req/min |
| Event tracking           | IP address | 60 req/min |

Rate limit headers returned on every request:
```
X-RateLimit-Limit: 60
X-RateLimit-Remaining: 45
Retry-After: 60   (only on 429)
```

---



**How would you separate write-heavy workloads from read-heavy workloads?**

Event writes (`POST /events/track`) are decoupled from reads by routing through SQS. The HTTP handler returns immediately after publishing; a background consumer handles DB writes asynchronously. Analytics reads hit `campaign_daily_metrics` (a pre-aggregated table, not raw logs) with a Redis cache layer in front. This means spikes in event ingestion have zero impact on analytics query latency.

**How would you design a pipeline to stream and aggregate campaign performance data in near real-time?**

The current SQS + consumer pattern gives ~20-second end-to-end latency (SQS long-poll window). For true near-real-time, replace SQS with Kafka — producers write to a topic, stream processors (Kafka Streams or Flink) aggregate over tumbling windows (e.g. 1-minute), and write rollups to a time-series store. Redis can also serve as a real-time counter layer, flushed to Postgres on a schedule.

**Would you pre-aggregate daily metrics? If so, how would you store and update them?**

Yes — `campaign_daily_metrics` stores one row per `(campaign_id, date)`. The consumer increments it atomically using `ON CONFLICT DO UPDATE SET clicks = clicks + EXCLUDED.clicks`. Weekly and monthly views are computed by `DATE_TRUNC + GROUP BY` over this table at query time, which is fast because there are at most 365 rows per campaign per year instead of millions of raw events.

**How would you structure your API to support exporting large datasets without timing out?**

Use cursor-based pagination (`WHERE id > last_seen_id ORDER BY id LIMIT n`) instead of OFFSET, which degrades at depth. For bulk exports, stream the Postgres result set directly to the HTTP response using `rows.Next()` + `json.NewEncoder(w)` in chunks, with `Transfer-Encoding: chunked`. For very large exports (>1M rows), generate the file asynchronously, upload to S3, and return a presigned download URL.

**How would you reduce cloud storage and compute costs for storing clickstream or impression logs?**

Partition `event_logs` by month (`PARTITION BY RANGE (occurred_at)`) and drop old partitions instead of running expensive `DELETE` sweeps. For cold data beyond the retention window, export partitions to S3 as Parquet (columnar, ~10x compression vs raw Postgres rows) and query via Athena. Hot data (last 30 days) stays in Postgres; warm data (30–90 days) stays in a cheaper RDS instance; cold data lives in S3.

---

## Future Improvements

- Replace SQS with Kafka for event replay, real-time streaming, and multi-consumer fan-out
- Add Prometheus metrics (event ingestion rate, queue lag, p99 API latency)
- OpenTelemetry distributed tracing across handler → SQS → consumer
- Data warehouse export (BigQuery / Snowflake) for long-term analytics
- WebSocket streaming for real-time dashboard updates
- Circuit breakers (Sony gobreaker) for SQS and Postgres dependencies
- Docker Compose setup for local development
- Database read replicas for analytics query isolation
