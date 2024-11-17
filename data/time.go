package data

import (
	"fmt"
	"strings"
	"time"
)

const DefaultTimeFormat = time.RFC3339

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

	parsedTime, err := time.ParseInLocation(timeFormat, strings.Trim(string(data), "\""), time.Local)
	if err != nil {
		return err
	}
	t.Time = parsedTime
	return nil
}

func (t Time) MarshalJSON() ([]byte, error) {
	return ([]byte)(fmt.Sprintf("\"%s\"", t.Time.Format(timeFormat))), nil
}
