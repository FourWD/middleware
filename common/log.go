package common

import (
	"github.com/FourWD/middleware/infra"
)

// Deprecated: use infra.AppLog.Event(label, data, requestID, opts...) directly
// and pass the required component/operation/log_kind metadata. This wrapper
// hides those fields and will be removed once all callers migrate.
// See CLAUDE.md "Logging Standard".
func Log(label string, logData map[string]interface{}, requestID string) {
	infra.AppLog.Event(label, logData, requestID, infra.WithCallerSkip(1))
}

// Deprecated: use infra.AppLog.EventWarn(label, data, requestID, opts...) directly.
// See CLAUDE.md "Logging Standard".
func LogWarning(label string, logData map[string]interface{}, requestID string) {
	infra.AppLog.EventWarn(label, logData, requestID, infra.WithCallerSkip(1))
}

// Deprecated: use infra.AppLog.EventError(err, label, data, requestID, opts...) directly.
// See CLAUDE.md "Logging Standard".
func LogError(label string, logData map[string]interface{}, requestID string, err ...error) {
	var e error
	if len(err) > 0 {
		e = err[0]
	}
	infra.AppLog.EventError(e, label, logData, requestID, infra.WithCallerSkip(1))
}
