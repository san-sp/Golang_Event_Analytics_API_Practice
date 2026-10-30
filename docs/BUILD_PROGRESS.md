# PulseTrack — Build Progress

## Project goal

Build a multi-tenant event ingestion and analytics REST API.

## Stack

- Go
- Gin
- MongoDB
- Official MongoDB Go driver
- Docker Compose
- Git

## Architecture decisions

- Modular monolith: one API service and one MongoDB database.
- Configuration comes from environment variables.
- Local development uses `.env`, which must not be committed.
- MongoDB uses a dedicated application user with limited permissions.
- Protected API operations must derive tenant identity from validated API keys.
- Automated test code is deferred until the end of the project, as requested.

## Completed and verified

- Initialized Go module `pulsetrack`.
- Created the initial package directory structure.
- Initialized Git.
- Added Gin HTTP server.
- Added configuration loading and port validation.
- Started MongoDB using Docker Compose with persistent storage.
- Created the `pulsetrack_app` MongoDB user.
- Verified `GET /healthz`.
- Verified `GET /readyz` returns HTTP 200 and database connected.
- Implemented tenant repository creation and retrieval methods.
- Verified formatting and successful compilation with `go build ./...`.
- Implemented API-key repository creation and lookup by hash.
- Verified formatting and successful compilation with `go build ./...`.
- Implemented event creation and tenant-scoped event retrieval.
- Applied bounded page sizes and deterministic event ordering.
- Verified formatting and successful compilation with `go build ./...`.
- Implemented cryptographically secure API-key generation.
- Added SHA-256 hashing before persistence.
- Added optional API-key expiration validation.
- Verified formatting and successful compilation with `go build ./...`.
- Key validation: Hashes the presented API key and looks up its metadata in MongoDB.
- Expiry and revocation checks: Rejects keys that have expired or have a revocation timestamp.
- Tenant status validation -Rejects keys belonging to disabled tenants and returns the stored tenant ID only after validation succeeds.
- Tenant-scoped revocation — one tenant cannot revoke another tenant's key.
- Duplicate revocation is prevented by the MongoDB filter.
- The service validates its inputs and wraps repository errors.
- gofmt and go build ./... pass.
- Simplify the revocation check
- Normalize the expiry timestamp
- Wire the repositories and services
- Wire the repositories in main.go

## Global step tracking

- Last completed implementation step: Step 5 — MongoDB setup and connection verification.
- Next available step: Step 6 — Build progress ledger.
- Step numbering is global and must remain sequential.
- Last completed implementation step: Step 11 — MongoDB index initialization and verification.
- Next available step: Step 12 — Tenant repository and persistence operations.
- Last completed implementation step: Step 12 — Tenant repository implementation and build verification.
- Next available step: Step 13 — API-key repository.
- Last completed implementation step: Step 13 — API-key repository implementation and build verification.
- Next available step: Step 14 — Event repository.
- Last completed implementation step: Step 14 — Event repository implementation and build verification.
- Next available step: Step 15 — Review repository layer and prepare application services.
  Step 16 — API-key service: generation and hashing
- Last completed implementation step: Step 16 — API-key generation and hash-only storage service.
- Next available step: Step 17 — API-key validation and revocation checks.
  Step 18 — Add API-key revocation
- Step 18 — API-key revocation is now implemented at both the repository and service layers.
  Step 19 — API-key service review
  Step 20 — API-key HTTP handlers
  Step 21 — Create the API-key HTTP handler
  Step 21 — Build the authentication middleware

## Endpoints currently available

- `GET /healthz`
- `GET /readyz`

## Index initialization

- Application-specific indexes have not yet been implemented.

## Tests

- Automated test code intentionally deferred until the end.
- Manual health and readiness checks passed.
- No automated test results claimed.

## Known limitations

- Event ingestion is not implemented.
- Tenant and API-key management are not implemented.
- Event querying and analytics are not implemented.
- Docker Compose currently runs MongoDB; API containerization remains pending.

## Next task

Implement API-key validation by hash and reject expired or revoked keys before allowing tenant-scoped requests.
