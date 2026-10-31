# PulseTrack – Go Event Analytics API – Part 2: Go Concurrency Engineering Extension

## Project Context

We are extending the existing **PulseTrack – Go Event Analytics API**.

GitHub repository:
https://github.com/san-sp/Golang_Event_Analytics_API_Practice

Existing technology stack:

- Go
- Gin
- MongoDB (official Go driver v2)
- Docker Compose
- Prometheus
- Grafana
- Alertmanager

The project follows a modular monolith architecture.

The original project was designed as a synchronous event ingestion and analytics API. Its original master prompt explicitly excluded queues and background workers.

**This phase is an intentional extension beyond the original scope. Do not silently change or discard the original architecture.**

## Main Objective

Introduce practical, production-minded Go concurrency concepts into PulseTrack.

The goal is to understand and demonstrate:

- Goroutines
- Channels
- Buffered and unbuffered channels
- Worker pools
- `sync.WaitGroup`
- `sync.Mutex`
- `sync.RWMutex`
- `sync.Once`
- `sync/atomic`
- Context cancellation
- Graceful shutdown
- Backpressure
- Concurrent processing
- Race-condition prevention
- Error propagation
- Resource management
- Concurrency testing

Do not introduce concurrency patterns merely to demonstrate language features. Every implementation must solve a real problem.

## Critical Working Rules

1. Inspect the complete existing repository before modifying anything.
2. Read the existing master prompt and `docs/BUILD_PROGRESS.md`.
3. Inspect the existing architecture, MongoDB repositories, services, handlers, middleware, configuration, and Docker setup.
4. Identify completed features, incomplete features, existing tests, and architectural constraints.
5. Do not rewrite working code unnecessarily.
6. Preserve existing API contracts unless a change is justified and documented.
7. Do not remove existing features.
8. Do not introduce PostgreSQL, Redis, RabbitMQ, Kafka, Kubernetes, or microservices.
9. Do not add dependencies without explaining why they are necessary.
10. Do not assume a feature exists just because a file or configuration entry exists. Verify its implementation.
11. Never claim tests passed unless you actually ran them and observed the result.
12. Do not commit or push changes unless I explicitly request it.

## Phase 1 – Repository Audit

Before coding:

- Inspect the repository.
- Understand the existing event ingestion workflow.
- Identify where synchronous processing currently occurs.
- Inspect MongoDB connection and repository behavior.
- Inspect the existing analytics aggregation workflow.
- Review existing monitoring and metrics.
- Identify concurrency risks and opportunities.

Produce a concise report containing:

- Existing architecture.
- Completed functionality.
- Pending functionality.
- Existing tests and coverage gaps.
- Recommended concurrency design.
- Files likely to require modification.

Do not start implementation until the audit is complete.

## Phase 2 – Design a Bounded Worker Pool

Design an optional asynchronous event-processing capability.

Requirements:

- Use a bounded buffered channel for incoming work.
- Use a configurable number of worker goroutines.
- Use `sync.WaitGroup` to coordinate worker lifecycle.
- Use `context.Context` for cancellation.
- Support graceful shutdown.
- Prevent unbounded memory growth.
- Implement explicit backpressure behavior.
- Ensure workers do not leak.
- Handle processing errors properly.
- Avoid data races.
- Ensure the MongoDB client is used safely.

The worker pool must have a clear lifecycle:

1. Initialize.
2. Start workers.
3. Accept work.
4. Process events.
5. Handle failures.
6. Stop accepting new work.
7. Drain or explicitly reject pending work according to the documented shutdown policy.
8. Wait for workers to finish.
9. Shut down cleanly.

### Important reliability requirement

Do not acknowledge an event as successfully persisted merely because it entered an in-memory channel.

Document the delivery semantics clearly.

Consider what happens when:

- The channel is full.
- MongoDB becomes unavailable.
- A worker crashes or returns an error.
- The API receives a large burst of requests.
- The application receives SIGTERM.
- The application shuts down while events remain queued.

Do not introduce a false guarantee of durable delivery. If the in-memory queue cannot guarantee persistence across process crashes, explicitly document that limitation.

## Phase 3 – Integrate With Event Ingestion

Evaluate how the worker pool should integrate with the existing event ingestion endpoint.

Consider two processing modes:

- Synchronous ingestion: preserve the existing behavior.
- Asynchronous ingestion: use the worker pool with clearly documented acceptance and persistence semantics.

Do not change the default behavior without justification.

If asynchronous ingestion requires an API contract change, explain the change and update the relevant documentation and tests.

Keep tenant isolation and API-key authentication intact.

## Phase 4 – Concurrency Safety

Review the entire implementation for:

- Concurrent map access.
- Shared mutable state.
- Race conditions.
- Unsafe shutdown sequences.
- Goroutine leaks.
- Channel send/close races.
- Deadlocks.
- Resource leaks.
- MongoDB session or transaction misuse.

Use mutexes only where shared mutable state genuinely requires synchronization.

Use atomic operations only where appropriate.

Do not add mutexes or channels unnecessarily.

## Phase 5 – Testing

Write meaningful tests for:

- Worker pool startup and shutdown.
- Multiple workers processing jobs concurrently.
- Bounded queue behavior.
- Backpressure.
- Context cancellation.
- Graceful shutdown.
- Processing errors.
- MongoDB failures.
- Concurrent event ingestion.
- Tenant isolation under concurrent requests.
- Goroutine lifecycle and leak prevention.

Use deterministic synchronization in tests rather than arbitrary sleeps wherever possible.

Run:

`go test ./... -count=1`

Also run:

`go test -race ./...`

Run targeted tests before the full suite.

If integration tests require MongoDB, document and execute the required setup.

Never hide or skip failing tests just to achieve a green test result.

## Phase 6 – Observability

Extend the existing monitoring where useful.

Consider metrics for:

- Queue depth.
- Queue capacity.
- Accepted events.
- Rejected events.
- Successfully processed events.
- Failed events.
- Processing duration.
- Active workers.
- Worker shutdown duration.

Integrate with the existing Prometheus configuration where appropriate.

Do not create duplicate metrics or break existing dashboards.

## Phase 7 – Documentation

Update the project documentation to explain:

- Why concurrency was introduced.
- The worker-pool architecture.
- Goroutine lifecycle.
- Channel behavior.
- Backpressure strategy.
- Error handling.
- Shutdown behavior.
- Delivery guarantees and limitations.
- Configuration options.
- How to run tests.
- How to run the race detector.
- How to observe the worker pool using Prometheus and Grafana.

Add architecture diagrams where useful.

Update `docs/BUILD_PROGRESS.md` with completed and pending work.

## Phase 8 – Interview Preparation

For every major implementation, explain:

1. What problem does this solve?
2. Why did we choose this concurrency pattern?
3. Why use a buffered channel?
4. Why use a WaitGroup?
5. When would a mutex be necessary?
6. How does context cancellation work?
7. How do we prevent goroutine leaks?
8. What happens when the queue becomes full?
9. What happens if MongoDB goes down?
10. How does graceful shutdown work?
11. What are the limitations of an in-memory worker pool?
12. How would this design change with a durable message broker?

Provide practical Go interview questions based on the actual implementation.

I must be able to explain every important line of code confidently.

## Development Workflow

Work incrementally.

For every task:

1. Explain the problem.
2. Explain the proposed solution.
3. Identify exact files to modify.
4. Show complete code for manageable files or precise patches for larger files.
5. Explain important code decisions.
6. Run targeted tests.
7. Fix failures before proceeding.
8. Update the progress documentation.

Do not dump the entire implementation in one response.

Do not move to the next phase until the current phase is verified.

## Final Goal

At the end of this extension, PulseTrack should demonstrate practical Go concurrency engineering while retaining its original functionality and architectural identity.

The implementation must be technically sound, testable, observable, documented, and something I can confidently explain in a real Golang developer interview.
