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
	zero := 0
	i := 1 / zero
	_ = i
	os.Exit(0)
}
