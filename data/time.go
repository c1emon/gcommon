package data

import (
	"fmt"
	"time"
)

const DefaultTimeFormat = "2006-01-02 15:04:05"

var timeFormat = DefaultTimeFormat

func SetTimeFormat(format string) {
	timeFormat = format
}

func GetTimeFormat() string {
	return timeFormat
}

type Time struct {
	time.Time
}

func (t *Time) UnmarshalJSON(data []byte) error {

	parsedTime, err := time.ParseInLocation(timeFormat, string(data), time.Local)
	if err != nil {
		return err
	}
	t.Time = parsedTime
	return nil
}

func (t Time) MarshalJSON() ([]byte, error) {
	return ([]byte)(fmt.Sprintf("\"%s\"", t.Time.Format(timeFormat))), nil
}
