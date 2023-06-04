package common

import (
	"fmt"
	"time"
)

func SetThailandTimezone() {
	location, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		fmt.Println("Error loading timezone:", err)
		return
	}

	// Set the default timezone for the application
	time.Local = location
}
