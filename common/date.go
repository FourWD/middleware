package common

import (
	"time"
)

func NilDate() time.Time {
	dateString := "1900-01-01"
	date, _ := time.Parse("2006-01-02", dateString)
	return date
}