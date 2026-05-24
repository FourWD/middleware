package infra

import (
	"context"
	"strings"
)

// helperCallerSkip lets the structured-log "source" field point at the
// caller of these helpers, not at log_helpers.go itself. Bumping this is
// needed because: user → helper → Logger.* method → log → addSource adds
// one extra frame compared to a direct Logger.* method call.
const helperCallerSkip = 1

// LogDBError emits a DB operation failure with full taxonomy. Operation is
// the verb (create, update, delete, lookup); table identifies the relation;
// recordID is the row's primary key if known.
//
// The label is auto-built as DB_<OPERATION>_FAILURE (UPPER_SNAKE), so
// callers do not need to construct one.
//
// Example:
//
//	if err := db.Create(&user).Error; err != nil {
//	    infra.LogDBError(ctx, err, "create", "users", user.ID)
//	    return err
//	}
func LogDBError(ctx context.Context, err error, operation, table, recordID string) {
	logger := LoggerFromContext(ctx)
	if logger == nil {
		return
	}
	label := "DB_" + strings.ToUpper(operation) + "_FAILURE"
	logger.ErrorCtx(ctx, err, label, nil,
		WithComponent(ComponentDB),
		WithOperation(operation),
		WithLogKind(LogKindError),
		WithField("table", table),
		WithField("record_id", recordID),
		WithCallerSkip(helperCallerSkip),
	)
}

// LogFirebaseError emits a Firebase operation failure with full taxonomy.
// Operation is the verb (subscribe, unsubscribe, send, broadcast, batch_update);
// resource identifies the target (topic name, document path, or "" if N/A).
// The user/device token MUST NOT be passed as resource — it is in the PII
// deny-list. The label is auto-built as FIREBASE_<OPERATION>_FAILURE.
//
// Example:
//
//	if _, err := FirebaseMessageClient.SubscribeToTopic(ctx, tokens, topic); err != nil {
//	    infra.LogFirebaseError(ctx, err, "subscribe", topic)
//	    return err
//	}
func LogFirebaseError(ctx context.Context, err error, operation, resource string) {
	logger := LoggerFromContext(ctx)
	if logger == nil {
		return
	}
	label := "FIREBASE_" + strings.ToUpper(operation) + "_FAILURE"
	logger.ErrorCtx(ctx, err, label, nil,
		WithComponent(ComponentFirebase),
		WithOperation(operation),
		WithLogKind(LogKindError),
		WithField("resource", resource),
		WithCallerSkip(helperCallerSkip),
	)
}

// LogAuthSecurity emits an auth-related security event at warn level (use
// for things like "invalid signature", "blacklisted token", "permission
// denied"). The label is auto-built as AUTH_<OPERATION>_FAILURE and tagged
// LogKindSecurity so SIEM rules can pick it up separately from generic
// errors. Pass reason for non-error context (e.g. "blacklisted",
// "expired"); pass err for actual failures.
//
// Example:
//
//	if claims == nil {
//	    infra.LogAuthSecurity(ctx, nil, "read_session", "no claims attached")
//	    return ErrUnauthorized
//	}
func LogAuthSecurity(ctx context.Context, err error, operation, reason string) {
	logger := LoggerFromContext(ctx)
	if logger == nil {
		return
	}
	label := "AUTH_" + strings.ToUpper(operation) + "_FAILURE"
	if err != nil {
		logger.ErrorCtx(ctx, err, label, nil,
			WithComponent(ComponentAuth),
			WithOperation(operation),
			WithLogKind(LogKindSecurity),
			WithField("reason", reason),
			WithCallerSkip(helperCallerSkip),
		)
		return
	}
	logger.WarnCtx(ctx, label, nil,
		WithComponent(ComponentAuth),
		WithOperation(operation),
		WithLogKind(LogKindSecurity),
		WithField("reason", reason),
		WithCallerSkip(helperCallerSkip),
	)
}

// LogPubSubError emits a Google Pub/Sub failure with full taxonomy.
// Operation is the verb (publish, subscribe, receive); topic is the topic
// name. The label is auto-built as PUBSUB_<OPERATION>_FAILURE.
//
// Example:
//
//	if _, err := publisher.Publish(ctx, msg).Get(ctx); err != nil {
//	    infra.LogPubSubError(ctx, err, "publish", topicName)
//	    return err
//	}
func LogPubSubError(ctx context.Context, err error, operation, topic string) {
	logger := LoggerFromContext(ctx)
	if logger == nil {
		return
	}
	label := "PUBSUB_" + strings.ToUpper(operation) + "_FAILURE"
	logger.ErrorCtx(ctx, err, label, nil,
		WithComponent(ComponentPubSub),
		WithOperation(operation),
		WithLogKind(LogKindError),
		WithField("topic", topic),
		WithCallerSkip(helperCallerSkip),
	)
}

// LogStorageError emits a Google Cloud Storage failure with full taxonomy.
// Operation is the verb (upload, download, delete, list); bucket and path
// identify the GCS object. The label is auto-built as
// STORAGE_<OPERATION>_FAILURE.
//
// Example:
//
//	if _, err := writer.Close(); err != nil {
//	    infra.LogStorageError(ctx, err, "upload", bucket, objectPath)
//	    return err
//	}
func LogStorageError(ctx context.Context, err error, operation, bucket, path string) {
	logger := LoggerFromContext(ctx)
	if logger == nil {
		return
	}
	label := "STORAGE_" + strings.ToUpper(operation) + "_FAILURE"
	logger.ErrorCtx(ctx, err, label, nil,
		WithComponent(ComponentStorage),
		WithOperation(operation),
		WithLogKind(LogKindError),
		WithField("bucket", bucket),
		WithField("path", path),
		WithCallerSkip(helperCallerSkip),
	)
}

// LogCritical emits a critical failure — reserved for "should never happen"
// events: panic recovery, data corruption, cryptography failure, invariant
// violation. Stamps severity=critical so alerting rules can route to
// pager/on-call separately from generic errors. component identifies the
// subsystem; operation is the verb in snake_case.
//
// Example:
//
//	defer func() {
//	    if r := recover(); r != nil {
//	        infra.LogCritical(ctx, fmt.Errorf("panic: %v", r),
//	            "WORKER_PANIC", infra.ComponentApp, "process_job")
//	    }
//	}()
func LogCritical(ctx context.Context, err error, label, component, operation string) {
	logger := LoggerFromContext(ctx)
	if logger == nil {
		return
	}
	logger.ErrorCtx(ctx, err, label, nil,
		WithComponent(component),
		WithOperation(operation),
		WithLogKind(LogKindError),
		WithField("severity", "critical"),
		WithCallerSkip(helperCallerSkip),
	)
}

// LogHTTPClientError emits an outbound HTTP failure with full taxonomy.
// Stage identifies which step of the request failed (encode, build, execute,
// status, read, unmarshal, …); the label is auto-built as
// HTTP_<STAGE>_FAILURE. Method is GET/POST/…; url is the target; status is
// the response status code (0 if the request never completed).
//
// Example:
//
//	resp, err := client.Do(req)
//	if err != nil {
//	    infra.LogHTTPClientError(ctx, err, "execute", req.Method, req.URL.String(), 0)
//	    return err
//	}
//	if resp.StatusCode >= 500 {
//	    infra.LogHTTPClientError(ctx, fmt.Errorf("upstream 5xx"),
//	        "status", req.Method, req.URL.String(), resp.StatusCode)
//	}
func LogHTTPClientError(ctx context.Context, err error, stage, method, url string, status int) {
	logger := LoggerFromContext(ctx)
	if logger == nil {
		return
	}
	label := "HTTP_" + strings.ToUpper(stage) + "_FAILURE"
	logger.ErrorCtx(ctx, err, label, nil,
		WithComponent(ComponentHTTPClient),
		WithOperation(strings.ToLower(method)),
		WithLogKind(LogKindError),
		WithField("stage", stage),
		WithField("url", url),
		WithField("status", status),
		WithCallerSkip(helperCallerSkip),
	)
}

// LogBusinessEvent emits an info-level business event with required
// metadata. Use when a service repo wants a one-line audit entry for a
// domain action where there is a request context in scope.
//
// Example:
//
//	infra.LogBusinessEvent(ctx, "USER_CREATED",
//	    infra.ComponentHandler, "create_user",
//	    map[string]any{"user_id": newUser.ID})
func LogBusinessEvent(ctx context.Context, label, component, operation string, data map[string]any) {
	logger := LoggerFromContext(ctx)
	if logger == nil {
		return
	}
	logger.EventCtx(ctx, label, data,
		WithComponent(component),
		WithOperation(operation),
		WithLogKind(LogKindBusiness),
		WithCallerSkip(helperCallerSkip),
	)
}

// LogLifecycle is the package-level shortcut for AppLog.LifecycleEvent.
// Use during boot/shutdown/connect/disconnect where request scope does not
// exist. Auto-stamps LogKindLifecycle.
//
// Example:
//
//	func startWorker(name string) {
//	    go runWorker(name)
//	    infra.LogLifecycle("WORKER_STARTED", infra.ComponentApp, "worker_start",
//	        map[string]any{"worker": name})
//	}
func LogLifecycle(label, component, operation string, data map[string]any) {
	if AppLog == nil {
		return
	}
	AppLog.LifecycleEvent(label, data,
		WithComponent(component),
		WithOperation(operation),
		WithCallerSkip(helperCallerSkip),
	)
}

// Deprecated: use LogLifecycle — same semantics, idiomatic name. The Infra
// kind was renamed to Lifecycle for clarity (boot/shutdown is one kind of
// lifecycle event, not a separate "infrastructure" category).
//
// The body is intentionally inlined rather than delegating to LogLifecycle:
// an extra Go-function frame would shift the structured-log `source` field
// to log_helpers.go. Inlining keeps both helpers at the same call depth.
func LogInfraEvent(label, component, operation string, data map[string]any) {
	if AppLog == nil {
		return
	}
	AppLog.LifecycleEvent(label, data,
		WithComponent(component),
		WithOperation(operation),
		WithCallerSkip(helperCallerSkip),
	)
}
