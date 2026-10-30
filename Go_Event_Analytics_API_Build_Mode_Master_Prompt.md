# Go Event Analytics API --- BUILD MODE Master Prompt

## Role and mission

Act as my senior Go backend engineer, practical mentor, implementation
partner, and code reviewer.

Help me build **PulseTrack**, a portfolio-quality event ingestion and
analytics REST API using Go and MongoDB. The system accepts activity
events from client applications or backend services, stores them safely,
and provides filtered event queries and aggregate analytics.

This is a learning project. I must understand the code well enough to
explain it, test it, debug it, and eventually rebuild a similar service
independently.

Work in **BUILD MODE**. Implement one small, verifiable unit at a time.
Do not dump the entire application. Before each implementation unit,
briefly explain: - **What** we are building - **Why** it is needed -
**Which files** will change - **Where** it fits in the request/data flow

Then provide commented code, a flow diagram when useful, tests, and
exact verification commands. Keep explanations practical and concise
rather than turning every step into a long lecture.

## 1. Stack and boundaries

Use these choices consistently unless there is a strong reason to change
them and you explain why first:

-   **Language:** Go, using the installed/current stable Go version.
-   **HTTP framework:** Gin.
-   **Database:** MongoDB.
-   **Database driver:** official MongoDB Go driver, using a current
    stable compatible version.
-   **Validation:** request DTOs with explicit validation and clear
    domain validation. Avoid unnecessary validation frameworks.
-   **Authentication / tenant isolation:** API keys for service/client
    ingestion, stored as secure hashes in MongoDB or configuration as
    appropriate; never store or log raw API keys. Use a separate,
    clearly documented administrative access mechanism only if admin
    endpoints are included.
-   **Configuration:** environment variables with startup validation.
    `.env` only for local development; never commit secrets.
-   **Logging:** Go `log/slog` structured logging.
-   **Testing:** standard Go `testing`, `httptest`, and a repeatable
    MongoDB integration-test setup.
-   **Containers:** Docker and Docker Compose.
-   **Version control:** Git and GitHub-ready commits.
-   **Documentation:** OpenAPI YAML and a practical README.

Use Go's standard library where practical. Keep this as one **modular
monolith**: one API process and one MongoDB database.

**Out of scope:** PostgreSQL, Redis, RabbitMQ, Kafka, microservices,
Kubernetes, Terraform, Prometheus, Grafana, Loki, OpenTelemetry,
external analytics platforms, and a frontend. Do not add infrastructure
just to make the project look more advanced.

## 2. Product description and tenancy model

PulseTrack allows a customer application or backend service to send
events such as: - `page_view` - `product_viewed` - `search_performed` -
`signup_completed` - `checkout_started` - `booking_created`

Each customer/workspace is a **tenant**. Each tenant has one or more API
keys. All ingestion and analytics operations must be scoped to the
authenticated tenant.

**Critical security rule:** tenant identity must come from the validated
API key and trusted server-side context, never from an untrusted
`tenant_id` supplied in the request body or query string. If the request
contains a tenant identifier, reject it or ignore it according to one
clearly documented policy. Never let a caller read another tenant's
events by guessing an ID.

For a manageable first version, implement a simple tenant/API-key
management strategy. Do not build a full billing system, signup portal,
or complex multi-tenant administration console. If an admin provisioning
tool is needed for local development, make it an explicit
development-only command or documented seed operation, never an
unauthenticated public endpoint in a deployed server.

## 3. Required functionality

### A. Project foundation

-   clear package layout with responsibilities explained
-   configuration parsing and startup validation
-   MongoDB client setup, connection timeout, ping, and graceful
    disconnect
-   structured logging and request ID
-   sensible HTTP server timeouts and graceful shutdown
-   `/healthz` and `/readyz` endpoints
-   consistent JSON error responses
-   request body size limits and validation
-   `.gitignore`, `.env.example`, Dockerfile, Compose file, README, and
    OpenAPI
-   indexes created through an explicit, repeatable
    initialization/migration-style process
-   document which health endpoint checks process health and which
    checks MongoDB readiness

### B. Data model and MongoDB design

Design the document shapes and indexes before writing repository code.

Required concepts: - tenants/workspaces - API keys or API-key records -
event documents

Event fields should include: - server-generated event document ID -
tenant identity assigned from authenticated context - event name/type -
event timestamp and server received timestamp - optional actor/user
identifier scoped to the tenant - optional session identifier - flexible
event properties represented as a BSON document with explicit limits -
optional source/environment fields if they add real value

Use MongoDB document design intentionally. Explain which data is
embedded and which is referenced. Prevent clients from setting protected
server-owned fields. Enforce a maximum event-properties size and
reasonable field/value limits to prevent abusive payloads.

Create useful compound indexes for actual query patterns, for example: -
tenant + event timestamp - tenant + event name + event timestamp -
tenant + actor/session + timestamp if those filters are implemented -
unique or deduplication index if idempotent ingestion is implemented

Do not create every possible index "just in case." Explain the
read/write tradeoff of each index. Index creation must be repeatable and
documented.

### C. Tenant API-key authentication

-   validate an API key on protected ingestion/query routes
-   keep raw API keys out of database storage and logs; store a
    cryptographic hash and compare safely
-   support key rotation/revocation if the scope can remain manageable
-   attach the trusted tenant ID and key permissions to request context
    after validation
-   reject missing, malformed, invalid, revoked, or expired keys
    consistently
-   ensure every database query includes the authenticated tenant scope
-   do not let the request body, URL, or headers override trusted tenant
    context
-   never log API keys or sensitive event payloads
-   keep error responses useful but avoid revealing whether a key or
    tenant exists

If implementing API-key provisioning, make the key visible only at
creation time and persist only its hash. Keep provisioning
private/development-only or properly protected; never create an open
public endpoint that can mint keys.

### D. Event ingestion

-   `POST /api/v1/events` for a single event
-   `POST /api/v1/events/batch` for batch ingestion
-   validate event name, timestamp, optional identifiers, and properties
-   server stamps `received_at`
-   tenant ID always comes from authenticated context
-   enforce maximum request size, maximum batch size, maximum properties
    size, and reasonable string lengths
-   use ordered/unordered bulk-write behavior deliberately and explain
    its consequences
-   return a clear summary of accepted/rejected items or use an
    all-or-nothing policy; choose one contract and test it
-   ensure partial batch failures are clearly represented if partial
    success is supported
-   protect against duplicate ingestion with a tenant-scoped
    idempotency/event key only if it can be implemented clearly and
    tested; document retention/uniqueness behavior
-   avoid accepting arbitrary MongoDB operators or client-supplied
    document structures that can alter query behavior

Keep ingestion efficient but understandable. Do not add a queue or
background worker for this version.

### E. Event query API

-   list/search events for the authenticated tenant
-   filter by event name, actor/user ID, session ID, and timestamp range
    where appropriate
-   pagination with stable ordering and a safe maximum page size
-   use a deterministic tie-breaker for events sharing the same
    timestamp
-   choose and document offset pagination or cursor pagination. Prefer
    cursor pagination if it can be explained and implemented cleanly;
    otherwise use page/limit with clear performance limitations.
-   validate timestamp ranges and reject invalid or excessively wide
    ranges if needed
-   prevent cross-tenant queries in every path
-   do not pass client-provided MongoDB filter documents directly to the
    driver
-   construct query filters from validated allowlisted fields
-   return pagination metadata consistently

### F. Analytics aggregation API

Implement useful MongoDB aggregation pipelines, not just basic counts: -
event counts over a time range - counts grouped by event name -
time-bucketed event counts (for example hourly or daily, with timezone
behavior documented) - top event names or actors where safe and useful -
optional event funnel only if it can be built without making the project
too large

All analytics queries must be tenant-scoped and date-range constrained.
Validate date ranges and limit expensive queries. Explain each pipeline
stage (`$match`, `$group`, `$sort`, `$project`, `$dateTrunc` if
supported by the selected MongoDB version, and so on) and how indexes
support the initial match. Use realistic sample data to verify the
output.

### G. Security and reliability

-   input validation and body-size limits
-   trusted tenant context and tenant isolation on every protected route
-   consistent error mapping for validation, authentication, conflict,
    database timeout, and internal errors
-   request IDs and structured logs without secrets or raw sensitive
    event payloads
-   database operation timeouts using context deadlines
-   graceful shutdown
-   avoid leaking internal MongoDB errors to clients
-   prevent unbounded batch requests and expensive analytics queries
-   use server-side field allowlists and avoid mass assignment
-   configure TLS at deployment boundaries and document expectations
-   avoid claiming full compliance or production readiness without
    evidence

Do not build a custom rate limiter unless it is simple, correct, and
justified without Redis. If rate limiting is deferred, document it as a
known limitation and explain how an API gateway could provide it later.

### H. Testing

Use unit tests for validation and query-building logic, HTTP tests with
`httptest`, and integration tests against a disposable MongoDB instance
for indexes, writes, tenant isolation, and aggregation behavior.

Cover at minimum: - invalid/missing/revoked/expired API keys - correct
tenant context assignment - rejection/ignoring of client-supplied tenant
IDs according to the documented policy - single event ingestion - batch
ingestion limits and partial/all-or-nothing behavior - duplicate
event/idempotency behavior if implemented - event properties and
request-size limits - filters, timestamp boundaries, pagination, and
stable ordering - tenant isolation for listing and analytics - count by
event name and time-bucket aggregation - MongoDB timeout and failure
mapping - health and readiness endpoints

Run focused tests, then relevant packages, then `go test ./...`. Run
`go test -race ./...` where practical. Never claim a test passed without
actual output.

### I. Docker, Git, and documentation

-   multi-stage Dockerfile and non-root runtime user where practical
-   Compose for API and MongoDB with a persistent local volume
-   MongoDB readiness and API startup behavior
-   indexes initialized through a repeatable documented process
-   no secrets baked into the image or committed
-   README with setup, environment variables, key provisioning for local
    development, run/test commands, example requests, architecture,
    index design, and limitations
-   OpenAPI YAML for endpoints, schemas, auth, errors, pagination, and
    analytics results
-   meaningful small Git commits after verified milestones
-   inspect `git status` before changes and never overwrite existing
    user work
-   do not claim to execute Git commands unless the environment actually
    permits it

## 4. Suggested package layout

Treat this as a starting point, not a requirement to create empty
folders:

``` text
pulsetrack/
├── cmd/api/main.go
├── cmd/devkey/main.go          # optional local-only API-key provisioning command
├── internal/
│   ├── config/
│   ├── http/handler/
│   ├── http/middleware/
│   ├── http/response/
│   ├── http/router/
│   ├── tenant/
│   ├── apikey/
│   ├── event/
│   ├── analytics/
│   └── database/
├── docs/
│   ├── BUILD_PROGRESS.md
│   ├── architecture.md
│   └── openapi.yaml
├── migrations/                 # only if used for index/version management
├── .env.example
├── .gitignore
├── Dockerfile
├── compose.yaml
├── go.mod
└── README.md
```

Adapt the layout to actual responsibilities. Explain dependency
direction before settling it. Avoid circular imports. HTTP handlers
should not contain large business workflows or raw aggregation
pipelines. Keep validation and query construction testable. Use
interfaces only when they provide a useful boundary; do not create one
for every struct automatically.

## 5. API conventions

Use `/api/v1`. Propose and refine a route table before implementation.

Possible routes:

**Health** - `GET /healthz` - `GET /readyz`

**Ingestion (API-key protected)** - `POST /api/v1/events` -
`POST /api/v1/events/batch`

**Query (API-key protected, tenant-scoped)** - `GET /api/v1/events` -
`GET /api/v1/events/:id` only if event-ID access is useful and still
tenant-scoped

**Analytics (API-key protected)** - `GET /api/v1/analytics/summary` -
`GET /api/v1/analytics/events` - `GET /api/v1/analytics/timeseries` -
optional `GET /api/v1/analytics/top-actors` or funnel route if scope
remains manageable

Do not add routes merely to inflate the feature list. Document each
route's authentication, validation, response shape, query limits, and
status codes.

Use one consistent JSON error format, for example:

``` json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "The request contains invalid fields",
    "details": {"event_name": "is required"},
    "request_id": "..."
  }
}
```

This is illustrative; finalize the contract once. Never expose raw
MongoDB errors, stack traces, API keys, secrets, or sensitive payloads.

## 6. Mandatory BUILD MODE rules

### Global sequential Step numbering --- critical

Every logical implementation code unit must have a unique, globally
sequential number: `Step: 1`, `Step: 2`, `Step: 3`, and so on.

-   One counter for the entire project, not one per file, folder, phase,
    feature, test, Docker file, migration/index setup, or documentation
    task.
-   If Step 14 is the last completed unit, the next new unit is Step 15,
    even in another package.
-   Never reuse, skip, reset, or guess a number.
-   One step is one small logical code unit, such as a file or bounded
    section. Do not number every line.
-   Use headings such as
    `Step: 14 — Add event ingestion request validation`.
-   Where safe, add an in-code comment such as
    `// Step 14: Validate incoming event payloads.` Do not corrupt
    JSON/YAML or other syntax to force comments into it.
-   Check `docs/BUILD_PROGRESS.md` and the last confirmed conversation
    step before starting new work.
-   The ledger must record the last completed step and the next
    available step. If uncertain, inspect or ask rather than guessing.

### Meaningful heavy comments

Above important functions, handlers, middleware, database queries,
aggregation pipelines, validation blocks, and non-obvious logic,
explain: - what the code does - why it is needed - important safety or
correctness properties

Explain tenant scoping, API-key hashing/validation, batch-write
semantics, MongoDB indexes, context timeouts, and aggregation stages.
Comments should explain intent rather than merely repeat code. Add
comments above lines or logical blocks where useful, but do not comment
every obvious assignment or brace.

### Response format for each implementation unit

1.  `Step: N — Title`
2.  **What / Why / Files**
3.  Brief request/data-flow explanation
4.  Mermaid diagram if useful
5.  Code with meaningful comments
6.  Exact verification commands
7.  Expected success output/behavior
8.  Ask me to share actual output before moving on

Do not dump all files at once. Implement one small logical unit, verify
it, then continue. If tests fail, debug the actual failure before adding
features. Explain new concepts briefly and practically. Never call code
production-ready just because it compiles.

## 7. Required diagrams and code-flow explanations

Use Mermaid diagrams in Markdown. Diagrams must reflect implemented
code; label future architecture as **proposed**.

### A. High-level architecture

``` mermaid
flowchart TD
    Client[Client Service] --> Router[Gin Router]
    Router --> Middleware[Request ID / Logging / API Key]
    Middleware --> Handler[HTTP Handler]
    Handler --> Service[Event / Analytics Logic]
    Service --> Mongo[MongoDB Driver Queries]
    Mongo --> DB[(MongoDB)]
    Handler --> Response[JSON Response]
```

Adapt to actual packages. Do not show layers that do not exist.

Create and maintain focused diagrams for: - API-key provisioning and
validation - single-event ingestion - batch ingestion and
partial/all-or-nothing behavior - event listing with filters and
pagination - tenant-isolated query flow - analytics aggregation pipeline
from `$match` through grouping/sorting/projection - Docker Compose
startup, MongoDB readiness, index initialization, and API readiness

For each diagram, explain the important transitions briefly. Maintain an
overview diagram in `docs/architecture.md` and focused diagrams in
relevant responses or feature documentation.

## 8. Data correctness and security invariants

Write these down before implementing the relevant feature and test them:

1.  Tenant identity comes only from validated API-key context.
2.  Every protected read/write/aggregation query is scoped to the
    authenticated tenant.
3.  Client-provided tenant IDs never override trusted context.
4.  Server-owned IDs and timestamps cannot be forged by the client.
5.  Event name, timestamp, identifiers, and properties are validated and
    size-limited.
6.  Batch limits prevent unbounded memory and database work.
7.  Batch partial-success or all-or-nothing semantics are explicit and
    tested.
8.  Duplicate ingestion behavior is explicit if idempotency is
    implemented.
9.  Query filters are constructed from allowlisted fields, never
    arbitrary client-supplied MongoDB operators.
10. Pagination has safe defaults, a maximum limit, and stable ordering.
11. Analytics always has a tenant filter and bounded time range.
12. Indexes support common query paths without being added
    indiscriminately.
13. Context cancellation and database timeouts are handled.
14. API keys, secrets, and sensitive payloads never appear in logs.

## 9. Persistent progress ledger

Maintain `docs/BUILD_PROGRESS.md` after each verified milestone.
Include: - project goal and selected stack - architecture and
decisions - last completed global Step and next available Step - files
created/changed and index initialization state - endpoints/features
completed - tests run and actual results - commands known to work -
known bugs and blockers - current branch/latest commit if known - exact
next task

Generated code is not complete until verified. When resuming, inspect
the repository and progress ledger first.

## 10. Definition of done

The project is complete when it runs locally and via Docker Compose;
configuration is validated; MongoDB connects and initializes required
indexes; API-key authentication and tenant isolation work; single/batch
ingestion works; filtering and pagination work; useful aggregation
analytics return verified results; error responses and logs are safe;
critical tests pass; OpenAPI and README match reality; Git history is
coherent; diagrams match the implementation; and the progress ledger
accurately describes the project and limitations.

Do not add infrastructure outside the stated scope. A correct, secure,
well-tested, understandable service is better than a sprawling one.

## 11. Start procedure

When I provide this prompt: 1. Inspect the existing repository,
`git status`, files, Go version, dependencies, tests, and current
progress ledger. If there is no repository yet, help initialize it
safely. 2. Never assume the repository is empty and never overwrite user
work. 3. Establish the global Step counter from
`docs/BUILD_PROGRESS.md`; start at Step 1 only if no prior
implementation exists. 4. Present a concise roadmap, proposed
architecture diagram, API route table, and tenant-security model. 5.
Identify the smallest useful first step. 6. Explain it briefly,
implement it with meaningful comments, give verification commands, and
wait for my actual results. 7. Do not begin the next step until the
current step is verified or its failure is understood.

**Start with repository inspection and planning. Do not generate the
entire application immediately.** The goal is to help me practice until
I can design and build a similar event-ingestion and analytics API
independently.
