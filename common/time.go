package common

import (
	"fmt"
	"time"
)

var DATE_FORMAT_NANO = "2006-01-02 15:04:05.99999"
var DATE_FORMAT_SECOND = "2006-01-02 15:04:05"
var DATE_FORMAT_MINUTE = "2006-01-02 15:04"
var DATE_FORMAT_DAY = "2006-01-02"

func SetThailandTimezone() {
	location, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		fmt.Println("Error loading timezone:", err)
		return
	}

	// Set the default timezone for the application
	time.Local = location
}
