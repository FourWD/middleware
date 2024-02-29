package common

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/FourWD/middleware/model"
	"github.com/spf13/viper"
)

func RunLatestVersionOnly(url string) {
	if os.Getenv("GAE_SERVICE") == "" {
		return
	}

	type Response struct {
		Status  int           `json:"status"`
		Message string        `json:"message"`
		Data    model.AppInfo `json:"data"`
	}

	var response Response
	jsonData := CallUrl(url)
	if err := json.Unmarshal([]byte(jsonData), &response); err != nil {
		fmt.Println("CallWakeUp", err.Error())
		fmt.Println("************************** Wake up ERROR! **************************")
		Terminate()
		return
	}

	if response.Status == 1 {
		if response.Data.AppVersion != viper.GetString("app_version") && viper.GetString("production") == "true" {
			Terminate()
		} else {
			fmt.Printf("************************** App Version: %s [%s][%s] Wake up OK! **************************\n", response.Data.AppVersion, ENV, os.Getenv("GAE_VERSION"))
		}
	}
}
