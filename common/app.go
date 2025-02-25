package common

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/FourWD/middleware/model"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
)

func RunLatestVersionOnly() {
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
	fields := map[string]interface{}{
		"wake_up_url": wakeUpUrl,
	}
	// log.Println("wakeUpUrl: ", wakeUpUrl)
	var response Response
	jsonData := CallUrl(wakeUpUrl)
	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		// fmt.Println("CallWakeUp", err.Error()) //
		// fmt.Println("************************** Wake up ERROR! **************************")
		fields["error"] = err.Error()
		Log("WakeUp", fields, false, "")

		Terminate()
		return
	}

	if response.Status == 1 {
		if response.Data.AppVersion != viper.GetString("app_version") {
			App.AppVersion = response.Data.AppVersion
			Terminate()
		} else {
			// fmt.Printf("************************** %s [%s] Version: [%s - %s] Wake up OK! **************************\n", App.GaeService, App.Env, App.AppVersion, App.GaeVersion)
			fields["app.gae_service"] = App.GaeService
			fields["app.env"] = App.Env
			fields["app.app_version"] = App.AppVersion
			fields["app.gae_version"] = App.GaeVersion
			Log("WakeUp", fields, true, "")
		}
	}
}
