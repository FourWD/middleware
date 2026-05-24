package infra

import (
	"sync"
	"time"

	"github.com/FourWD/middleware/kit"
)

const (
	DATE_FORMAT_NANO   = "2006-01-02 15:04:05.99999"
	DATE_FORMAT_SECOND = "2006-01-02 15:04:05"
	DATE_FORMAT_MINUTE = "2006-01-02 15:04"
	DATE_FORMAT_DAY    = "2006-01-02"
)

var bangkokLocation = sync.OnceValues(func() (*time.Location, error) {
	return time.LoadLocation("Asia/Bangkok")
})

func parseWithTimezone(strDateTime string, format string) (time.Time, error) {
	location, err := bangkokLocation()
	if err != nil {
		AppLog.EventError(err, "TIME_LOCATION_FAILURE", map[string]any{"input": strDateTime}, "",
			WithComponent(ComponentApp),
			WithOperation("load_location"),
			WithLogKind(LogKindError))
		return kit.NilDate(), err
	}

	parsedTime, err := time.ParseInLocation(format, strDateTime, location)
	if err != nil {
		AppLog.EventError(err, "TIME_PARSE_FAILURE", map[string]any{"input": strDateTime}, "",
			WithComponent(ComponentApp),
			WithOperation("parse_time"),
			WithLogKind(LogKindError))
		return kit.NilDate(), err
	}

	return parsedTime, nil
}

// StringToDateTime parses "YYYY-MM-DD HH:MM" in Asia/Bangkok timezone.
func StringToDateTime(strDateTime string) (time.Time, error) {
	return parseWithTimezone(strDateTime, DATE_FORMAT_MINUTE)
}

// StringToDate parses "YYYY-MM-DD" in Asia/Bangkok timezone.
func StringToDate(strDateTime string) (time.Time, error) {
	return parseWithTimezone(strDateTime, DATE_FORMAT_DAY)
}
