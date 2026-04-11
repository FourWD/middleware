package common

import (
	"time"
)

var DATE_FORMAT_NANO = "2006-01-02 15:04:05.99999"
var DATE_FORMAT_SECOND = "2006-01-02 15:04:05"
var DATE_FORMAT_MINUTE = "2006-01-02 15:04"
var DATE_FORMAT_DAY = "2006-01-02"

func SetThailandTimezone() {
	location, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		LogError("TIMEZONE_LOAD_ERROR", map[string]interface{}{"error": err.Error()}, "")
		return
	}

	time.Local = location
}

func UTCToThailandTime(t time.Time) time.Time {
	bangkokLocation, _ := time.LoadLocation("Asia/Bangkok")
	return t.In(bangkokLocation)
}

