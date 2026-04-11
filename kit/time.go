package kit

import "time"

var (
	DateFormatNano   = "2006-01-02 15:04:05.99999"
	DateFormatSecond = "2006-01-02 15:04:05"
	DateFormatMinute = "2006-01-02 15:04"
	DateFormatDay    = "2006-01-02"
)

var nowFunc = time.Now

func Now() time.Time {
	return nowFunc()
}

func LoadTimezone(name string) (*time.Location, error) {
	return time.LoadLocation(name)
}

func InTimezone(t time.Time, loc *time.Location) time.Time {
	return t.In(loc)
}

func CheckInTime(startHour, startMinute, endHour, endMinute int) bool {
	now := Now()
	startTime := time.Date(now.Year(), now.Month(), now.Day(), startHour, startMinute, 0, 0, now.Location())
	endTime := time.Date(now.Year(), now.Month(), now.Day(), endHour, endMinute, 0, 0, now.Location())
	return now.After(startTime) && now.Before(endTime)
}

func SameDate(t1, t2 time.Time) bool {
	y1, m1, d1 := t1.Date()
	y2, m2, d2 := t2.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

func ParseRFC3339(value string) (time.Time, error) {
	return time.Parse(time.RFC3339, value)
}

func ParseDateOnly(value string) (time.Time, error) {
	return time.Parse(time.DateOnly, value)
}

func BeginningOfDay(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, value.Location())
}

func EndOfDay(value time.Time) time.Time {
	return BeginningOfDay(value).AddDate(0, 0, 1).Add(-time.Nanosecond)
}

func NilDate() time.Time {
	date, _ := time.Parse("2006-01-02", "1900-01-01")
	return date
}
