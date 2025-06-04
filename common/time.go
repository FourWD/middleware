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

func UTCToThailandTime(t time.Time) time.Time {
	//fmt.Println("UTC Time:", t)
	bangkokLocation, _ := time.LoadLocation("Asia/Bangkok")
	//bangkokTime := t.In(bangkokLocation)
	//bangkokTime = bangkokTime.Round(0)
	//fmt.Println("Bangkok Time:", bangkokTime)

	return t.In(bangkokLocation)
}

func CheckInTime(startHour, startMinute, endHour, endMinute int) bool {
	// Get the current time
	now := time.Now()

	// Create time objects for the start and end times using the current date
	startTime := time.Date(now.Year(), now.Month(), now.Day(), startHour, startMinute, 0, 0, now.Location())
	endTime := time.Date(now.Year(), now.Month(), now.Day(), endHour, endMinute, 0, 0, now.Location())

	// Check if the current time is between the start and end times
	return now.After(startTime) && now.Before(endTime)
}

func SameDate(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// func RoundUpToMinute(t time.Time) time.Time {
// 	rounded := time.Date(
// 		t.Year(),
// 		t.Month(),
// 		t.Day(),
// 		t.Hour(),
// 		t.Minute(),
// 		0, // seconds
// 		0, // nanoseconds
// 		t.Location(),
// 	)

// 	if t.Second() > 0 || t.Nanosecond() > 0 {
// 		rounded = rounded.Add(time.Minute)
// 	}

// 	return rounded
// }

// func RoundDownToMinute(t time.Time) time.Time {
// 	rounded := time.Date(
// 		t.Year(),
// 		t.Month(),
// 		t.Day(),
// 		t.Hour(),
// 		t.Minute(),
// 		0, // seconds
// 		0, // nanoseconds
// 		t.Location(),
// 	)

// 	return rounded
// }
