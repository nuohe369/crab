package util

import "time"

const (
	DateFormat     = "2006-01-02"
	TimeFormat     = "15:04:05"
	DateTimeFormat = "2006-01-02 15:04:05"
)

// FormatDate formats date (year-month-day)
func FormatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(DateFormat)
}

// FormatTime formats time (hour:minute:second)
func FormatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(TimeFormat)
}

// FormatDateTime formats datetime (year-month-day hour:minute:second)
func FormatDateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(DateTimeFormat)
}

// FormatDatePtr formats date pointer
func FormatDatePtr(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format(DateFormat)
}

// FormatDateTimePtr formats datetime pointer
func FormatDateTimePtr(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format(DateTimeFormat)
}

// ParseDate parses date
func ParseDate(s string) (time.Time, error) {
	return time.ParseInLocation(DateFormat, s, time.Local)
}

// ParseDateTime parses datetime
func ParseDateTime(s string) (time.Time, error) {
	return time.ParseInLocation(DateTimeFormat, s, time.Local)
}

// StartOfDay returns the start of the day
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// EndOfDay returns the end of the day
func EndOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
}

// Now returns the current time
func Now() time.Time {
	return time.Now()
}

// Today returns today's date string
func Today() string {
	return FormatDate(time.Now())
}

// NowStr returns the current datetime string
func NowStr() string {
	return FormatDateTime(time.Now())
}
