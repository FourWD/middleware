package common

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/FourWD/middleware/model"
	"github.com/robfig/cron/v3"
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

	wakeUpUrl := fmt.Sprintf("https://%s-dot-%s.appspot.com/wakeup", App.GaeService, App.GaeProject)
	var response Response
	jsonData := CallUrl(wakeUpUrl)
	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		// fmt.Println("CallWakeUp", err.Error())
		fmt.Println("************************** Wake up ERROR! **************************")
		Terminate()
		return
	}

	if response.Status == 1 {
		if response.Data.AppVersion != App.AppVersion {
			Terminate()
		} else {
			fmt.Printf("************************** %s [%s] Version: [%s - %s] Wake up OK! **************************\n", App.GaeService, App.Env, App.AppVersion, App.GaeVersion)
		}
	}
}
