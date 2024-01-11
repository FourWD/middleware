package common

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/spf13/viper"
)

func CallWakeUp(url string) {
	fmt.Println("CallWakeUp URL: ", url)
	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading the response body:", err)
		return
	}

	type Success struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
	}

	var data Success
	if err := json.Unmarshal(body, &data); err != nil {
		fmt.Println("RunWakeUp Error: ", err, string(body))
		return
	}

	if data.Status == 1 {
		log.Println(getServiceName())
		log.Println("************************** Wake up: OK!")
	} else {
		log.Println("************************** Wake up: ERROR!")
	}

}

func getServiceName() string {
	if viper.GetString("production") == "true" {
		return fmt.Sprintf("Service name: %s, Current version: %s", os.Getenv("GAE_SERVICE"), os.Getenv("GAE_VERSION"))
	}
	return "Service name: [LOCAL] Engine Service"
}
