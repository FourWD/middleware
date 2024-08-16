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

	type Response struct {
		Status  int           `json:"status"`
		Message string        `json:"message"`
		Data    model.AppInfo `json:"data"`
	}

	var response Response

	c := cron.New()
	c.AddFunc("*/1 * * * *", func() {
		jsonData := CallUrl(App.WakeUpUrl)
		if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
			fmt.Println("CallWakeUp", err.Error())
			fmt.Println("************************** Wake up ERROR! **************************")
			Terminate()
			return
		}

		if response.Status == 1 {
			if response.Data.AppVersion != viper.GetString("app_version") {
				Terminate()
			} else {
				App.AppVersion = response.Data.AppVersion
				fmt.Printf("************************** %s [%s] Version: [%s - %s] Wake up OK! **************************\n", App.GaeService, App.Env, App.AppVersion, App.GaeVersion)
			}
		}
	})
	c.Start()
}
