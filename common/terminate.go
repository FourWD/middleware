package common

import (
	"os"
)

func Terminate() {
	LogError("TERMINATE", map[string]interface{}{
		"service":     App.GaeService,
		"env":         App.Env,
		"app_version": App.AppVersion,
		"gae_version": App.GaeVersion,
	}, "")
	os.Exit(1)
}
