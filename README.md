# middleware

Shared boilerplate for Fiber services. It provides:

- app bootstrap and shutdown lifecycle
- database, redis, mongo, firebase initialization
- optional cloud/integration clients for pubsub, storage, and mailgun
- jwt/token helpers and optional mongo-backed token blacklist
- standard middleware stack, health routes, metrics, tracing, and logging

This module is referenced from the root project via:

```go
replace boilerplate.go/middleware => ./middleware
```

## Dependency Model

`infra.NewApp()` initializes common dependencies from env and passes grouped dependencies into each service registrar through `infra.AppDeps`:

- `Runtime`: config, logger, shutdown hooks, heartbeat debug status
- `Data`: primary/secondary databases, redis, mongo
- `Security`: shared auth helpers such as `BlacklistStore`
- `Cloud`: firebase, pubsub, storage
- `Integrations`: mailgun

## Config Matrix

Core:

- `APP_NAME`, `APP_ENV`, `HTTP_ADDRESS`
- `DB_*`, optional `DB2_*`
- `REDIS_*`
- `JWT_*`, `AUTH_BCRYPT_COST`
- `MIGRATIONS_ENABLED`, `MIGRATIONS_PATH`

Mongo and JWT blacklist:

- `MONGO_ENABLED`, `MONGO_URI`, `MONGO_DATABASE`
- `JWT_BLACKLIST_ENABLED`

Firebase:

- `FIREBASE_CREDENTIALS`
- `FIREBASE_NOTIFICATION_CREDENTIALS`

Pub/Sub:

- `PUBSUB_ENABLED`
- `PUBSUB_PROJECT_ID`
- `PUBSUB_CREDENTIALS_FILE`

Storage:

- `STORAGE_ENABLED`
- `STORAGE_BUCKET`
- `STORAGE_CREDENTIALS_FILE`

Mailgun:

- `MAIL_ENABLED`
- `MAILGUN_DOMAIN`
- `MAILGUN_API_KEY`

## Health Checks

Use `infra.HealthCheck(infra.HealthCheckOptions{...})` to choose which dependencies are included in readiness checks. Built-in checks support:

- databases
- redis
- mongo
- firestore

You can append custom checks through `HealthCheckOptions.Checks`.

### Endpoints

- `GET /livez`
  Returns basic process liveness only. This should stay lightweight and should not depend on external systems.
- `GET /healthz`
  Returns a full readiness report with per-component status.
- `GET /readyz`
  Same readiness report as `/healthz`, intended for load balancers and orchestrators.

### Report Semantics

`HealthReport.Status` can be:

- `ok`: all required and optional checks passed
- `degraded`: at least one optional check failed
- `down`: at least one required check failed

Each component reports:

- `status`
  `ok`, `degraded`, `down`, or `skipped`
- `required`
  Whether failure of that dependency should mark the service unhealthy
- `duration_ms`
  Check duration in milliseconds
- `error`
  Present when the check fails

### Example

```json
{
  "success": true,
  "data": {
    "status": "degraded",
    "checked_at": "2026-03-17T10:00:00Z",
    "components": {
      "database": {
        "status": "ok",
        "required": true,
        "duration_ms": 12
      },
      "mongo": {
        "status": "degraded",
        "required": false,
        "duration_ms": 31,
        "error": "server selection timeout"
      }
    }
  }
}
```

### Recommended Usage

- Use `/livez` for process liveness probes.
- Use `/readyz` for Kubernetes readiness probes and traffic gating.
- Use `/healthz` for operator visibility and debugging.
- Prefer service-specific `FirestoreCheck` callbacks instead of generic Firestore listing, so health checks match the real document paths each service uses.
