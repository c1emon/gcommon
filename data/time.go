package data

import (
	"bytes"
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
	return bytes.NewBufferString(warpDoubleQuotes(t.Time.Format(timeFormat))).Bytes(), nil
}
