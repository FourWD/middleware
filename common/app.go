package common

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/FourWD/middleware/model"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
)

func runLatestVersionOnly() {
	if os.Getenv("GAE_SERVICE") == "" {
		return
	}

	handleRunLatestVersionOnly() // run after deploy
	c := cron.New()
	c.AddFunc("*/1 * * * *", func() {
		handleRunLatestVersionOnly()
	})
	c.Start()
}

func handleRunLatestVersionOnly() {
	type Response struct {
		Status  int           `json:"status"`
		Message string        `json:"message"`
		Data    model.AppInfo `json:"data"`
	}

	wakeUpUrl := fmt.Sprintf("https://%s-dot-%s.appspot.com/wake-up", App.GaeService, App.GaeProject)
	logData := map[string]interface{}{
		"wake_up_url": wakeUpUrl,
	}

	var response Response
	jsonData := CallUrl(wakeUpUrl)
	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		logData["error"] = err.Error()
		LogError("WakeUp", logData, "")

		Terminate()
		return
	}

	if response.Status == 1 {
		if response.Data.AppVersion != viper.GetString("app_version") {
			App.AppVersion = response.Data.AppVersion
			Terminate()
		} else {
			logData["app.gae_service"] = App.GaeService
			logData["app.env"] = App.Env
			logData["app.app_version"] = App.AppVersion
			logData["app.gae_version"] = App.GaeVersion
			Log("WakeUp", logData, "")
		}
	}
}
