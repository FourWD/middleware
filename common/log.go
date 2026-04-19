package common

import (
	"github.com/FourWD/middleware/infra"
)

func Log(label string, logData map[string]interface{}, requestID string) {
	infra.AppLog.Event(label, logData, requestID, infra.WithCallerSkip(1))
}

func LogWarning(label string, logData map[string]interface{}, requestID string) {
	infra.AppLog.EventWarn(label, logData, requestID, infra.WithCallerSkip(1))
}

func LogError(label string, logData map[string]interface{}, requestID string, err ...error) {
	var e error
	if len(err) > 0 {
		e = err[0]
	}
	infra.AppLog.EventError(e, label, logData, requestID, infra.WithCallerSkip(1))
}
