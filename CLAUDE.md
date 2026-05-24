# CLAUDE.md

## General Behavior

- Be concise and direct.
- Do not over-explain.
- Focus on producing working, clean code.

## Code Style

- Write clean, readable, and idiomatic code.
- Prefer simplicity over cleverness.
- Follow standard conventions of the language (e.g., Go, TypeScript).

## Comments Policy

- Write comments ONLY when necessary.
- Avoid obvious comments that restate the code.
- Use comments for:
  - Complex logic
  - Non-obvious decisions
  - Edge cases

- All comments MUST be written in English.
- Keep comments short and precise.

## Restrictions

- Do NOT add unnecessary explanations outside the code unless explicitly requested.
- Do NOT generate verbose documentation.

## Example

### Bad

```go
// This function adds two numbers
func Add(a, b int) int {
    return a + b
}
```

### Good

```go
func Add(a, b int) int {
    return a + b
}
```

### Acceptable (when needed)

```go
func CalculateDiscount(price float64, userType string) float64 {
    // Apply special rate for premium users due to business rule v2
    if userType == "premium" {
        return price * 0.8
    }
    return price
}
```

## Logging Standard

### Default API

- **Business events** (handlers, services, use-cases): `infra.AppLog.EventCtx(ctx, label, data, opts...)` and `WarnCtx` / `ErrorCtx`.
- **Lifecycle** (boot, shutdown, worker start/stop, client connect): `infra.AppLog.LifecycleEvent(label, data, opts...)` and `LifecycleWarn` / `LifecycleError`. Auto-stamps `WithLogKind(LogKindLifecycle)`.
- **Common patterns**: use helpers in `infra/log_helpers.go` — `LogDBError`, `LogHTTPClientError`, `LogFirebaseError`, `LogPubSubError`, `LogStorageError`, `LogAuthSecurity`, `LogCritical`, `LogBusinessEvent`, `LogLifecycle`.
- **Removed:** the generic Level methods `Info/Warn/Error/CriticalWarning/CriticalError` no longer exist on `*Logger`. Migrate to `LifecycleXxx` / `EventCtx` / typed helpers — see migration map in repo PR history if upgrading from a pre-removal middleware version.
- **Do NOT** use `common.Log/LogWarning/LogError` in new code. Deprecated bridge.

### Scoped logger override (tests, request-scoped)

- `infra.WithLogger(ctx, logger)` attaches a scoped logger to a context.
- `infra.LoggerFromContext(ctx)` returns the scoped logger, falling back to `infra.AppLog`.
- Test helper `infra.CaptureLogs(t)` swaps `infra.AppLog` for the duration of `t` and restores via `t.Cleanup`. NOT safe with `t.Parallel()` — for parallel scoping use `WithLogger(ctx, …)` instead.

### Error code extraction

- `*infra.AppError` automatically contributes `error_code` and `error_status` log attributes.
- Service-defined error types satisfying `infra.ErrorWithCode` (`Code() string`) automatically contribute `error_code`. This lets a service repo decorate its domain errors for log aggregation without depending on `*infra.AppError`.

### Label format

- `UPPER_SNAKE_CASE` always.
- Shape: `<DOMAIN>_<OPERATION>_<OUTCOME>` (outcome optional).
  - `DB_CREATE_FAILURE`, `OTP_VERIFY_SUCCESS`, `FIREBASE_BATCH_START`.
- No spaces, no camelCase, no leading verbs.

### Mandatory fields per call

Every `Event/EventCtx/WarnCtx/ErrorCtx` call MUST include all three:

- `WithComponent(infra.Component...)` — pick from the enumerated `Component*` constants in `infra/log_taxonomy.go`. Full list: `app`, `http`, `db`, `mongo`, `redis`, `auth`, `firebase`, `pubsub`, `storage`, `mail`, `http_client`, `cron`, `handler`, `repo`, `payment`, `otp`, `upload`.
- `WithOperation("verb_in_snake_case")` — e.g. `create_user`, `verify_otp`, `batch_update`.
- `WithLogKind(infra.LogKind...)` — pick from the enumerated `LogKind*` constants in `infra/log_taxonomy.go`. Full list: `request`, `business_event`, `lifecycle`, `error`, `security`, `diagnostic`, `startup`, `infrastructure`.

Use the constants — do not invent new component/kind strings ad-hoc. Add new constants to `log_taxonomy.go` if a genuinely new subsystem appears.

### Field naming convention

- All `WithField` keys use `snake_case`.
- Use the same key across all callsites for the same concept — dashboards filter by exact key.
- Canonical keys: `user_id`, `request_id`, `trace_id`, `span_id`, `topic`, `bucket`, `path`, `table`, `record_id`, `duration_ms`, `count`, `status`, `url`, `method`, `resource`, `reason`, `severity`.
- Durations: always `*_ms` suffix and `int64` (`time.Since(start).Milliseconds()`).
- Counts/sizes: `int` for bounded values, `int64` for unbounded (file sizes, byte counts).

### Level discipline

| Level    | Use when                                                                                                                    |
| -------- | --------------------------------------------------------------------------------------------------------------------------- |
| Info     | Normal business outcome. Use `EventCtx` / `LifecycleEvent`.                                                                 |
| Warn     | Recoverable, degraded mode, single retry. Use `WarnCtx` / `LifecycleWarn`.                                                  |
| Error    | Request fails, requires investigation. Use `ErrorCtx` / `LifecycleError` / typed helpers (`LogDBError`, etc.).               |
| Critical | Should-never-happen: panic recovery, data corruption, cryptography failure. Use `LogCritical` (stamps `severity=critical`). |

- Do NOT log Error for "not found" / expected validation failure. Use Info or skip.
- Do NOT log Warn for every retry. Only when retry threshold exceeded.

### Error field discipline

- Pass `err` as the positional argument to `ErrorCtx(ctx, err, ...)`.
- NEVER duplicate the error into the data map: `data["error"] = err.Error()` is redundant — the logger already extracts it.
- Use `WithField("table", "users")` for additional context, not formatted strings inside the message.

### PII deny-list — NEVER log raw values of

- JWT, OAuth, FCM, Firebase, API tokens.
- Passwords, encryption keys, salts.
- Mobile numbers, email addresses, NRIC, passport, credit card numbers, bank accounts.
- Full request/response bodies (request log already gates this — do not bypass).

### PII allow-list — OK to log

- `user_id` (UUID format only).
- `email_domain` (extracted, not the local part).
- `mobile_last_4` via `kit.MaskMobile` (`XXX-XXX-1234`).
- `request_id`, `trace_id`, `span_id`.

### No logging in

- Tight loops over user input (>100 iterations) — log start + summary instead.
- Per-row DB CRUD success — the HTTP request log already proves the outcome.
- Hot middleware paths where the value adds nothing diagnostic.

### Helpers

Prefer the typed helpers in `infra/log_helpers.go` over hand-rolled `ErrorCtx` calls. Each helper stamps the correct component/operation/log_kind so the caller writes one line instead of five.

| Helper | Domain | Auto-built label | Extra fields |
| --- | --- | --- | --- |
| `LogDBError(ctx, err, op, table, recordID)` | DB | `DB_<OP>_FAILURE` | `table`, `record_id` |
| `LogMongoError` *(if added)* | Mongo | — | — |
| `LogHTTPClientError(ctx, err, method, url, status)` | outbound HTTP | `HTTP_CLIENT_FAILURE` | `url`, `status` |
| `LogFirebaseError(ctx, err, op, resource)` | Firebase | `FIREBASE_<OP>_FAILURE` | `resource` |
| `LogAuthSecurity(ctx, err, op, reason)` | auth | `AUTH_<OP>_FAILURE` | `reason`. Logs at warn if err is nil, error if err is non-nil. Tagged `LogKindSecurity`. |
| `LogPubSubError(ctx, err, op, topic)` | Pub/Sub | `PUBSUB_<OP>_FAILURE` | `topic` |
| `LogStorageError(ctx, err, op, bucket, path)` | GCS | `STORAGE_<OP>_FAILURE` | `bucket`, `path` |
| `LogCritical(ctx, err, label, component, op)` | any | (caller supplies) | `severity=critical` — reserved for invariant violations, panic recovery, crypto failures. |
| `LogBusinessEvent(ctx, label, component, op, data)` | any | (caller supplies) | none — for info-level business events with a custom data map |
| `LogLifecycle(label, component, op, data)` | any | (caller supplies) | tagged `LogKindLifecycle` — for boot/shutdown/worker events with NO request context |

If your pattern doesn't fit any helper, call `EventCtx` / `ErrorCtx` directly with the mandatory three options — and consider whether the pattern is common enough to deserve a new helper.

### Logger package policy

- **`infra/` and `common/`**: use `infra.AppLog` (the project `*Logger`) or the helpers above.
- **`kit/`**: stdlib `log/slog` only. `kit` does not import `infra` (layering rule), so the typed `*Logger` is unavailable. Callers that need a logger pass `*slog.Logger` in — see `kit/terminate.go` for the canonical pattern.
- **Service repos**: use `infra.AppLog` via the helpers; for request-scoped logs use `EventCtx(ctx, …)` so the trace ID and request ID propagate automatically.

### Cost and sampling

Hosted log systems (Datadog, Loki, Sentry) charge per event. For high-volume call sites:

- **>1000 events/sec sustained**: sample (e.g. log 1 in N) or aggregate (log a summary every N seconds, not every event).
- **Tight loops over user input (>100 iterations)**: log start + summary instead of per-item.
- **Per-row DB CRUD success**: already proven by the HTTP request log — do not log a separate event.
- **Hot middleware paths**: log only when the value adds diagnostic signal (errors, security events) — skip happy-path noise.

A request log line costs ~1 KB. At 100 RPS sustained that's ~8 GB/day per service. Budget accordingly.

### Examples

#### Business event with request context

```go
infra.AppLog.EventCtx(ctx, "USER_CREATED", map[string]any{
    "user_id": user.ID,
    "source":  "web",
},
    infra.WithComponent(infra.ComponentHandler),
    infra.WithOperation("create_user"),
    infra.WithLogKind(infra.LogKindBusiness))
```

#### Error path via helper (5 lines → 1)

```go
if err := db.Create(&user).Error; err != nil {
    infra.LogDBError(ctx, err, "create", "users", user.ID)
    return err
}
```

#### Auth security event (no err — just a deny)

```go
if !permitted {
    infra.LogAuthSecurity(ctx, nil, "check_role", "role mismatch")
    return ErrForbidden
}
```

#### Lifecycle event (no request context)

```go
infra.LogLifecycle("CACHE_WARMED", infra.ComponentApp, "warm_cache",
    map[string]any{"item_count": count, "duration_ms": elapsedMs})
```

#### Critical "should never happen"

```go
defer func() {
    if r := recover(); r != nil {
        infra.LogCritical(ctx, fmt.Errorf("panic: %v", r),
            "WORKER_PANIC", infra.ComponentApp, "process_job")
    }
}()
```

#### Hand-rolled when no helper fits

```go
infra.AppLog.WarnCtx(ctx, "CACHE_MISS_RATIO_HIGH", map[string]any{
    "ratio":  ratio,
    "window": "1m",
},
    infra.WithComponent(infra.ComponentApp),
    infra.WithOperation("monitor_cache"),
    infra.WithLogKind(infra.LogKindDiagnostic))
```
