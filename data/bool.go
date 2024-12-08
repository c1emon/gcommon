package data

import (
	"bytes"
	"strconv"
)

var negativeWords = []string{"null", "nil", "none", "false", "no", "0", "bad", "negative"}

type LooseBool struct {
	bool
	raw string
}

func NewLooseBool(boolen bool) LooseBool {
	return LooseBool{bool: boolen}
}

func (b *LooseBool) Set(boolen bool) {
	b.bool = boolen
}

func (b LooseBool) ToBool() bool {
	return b.bool
}

func (b *LooseBool) UnmarshalJSON(data []byte) error {

	dataStr := string(bytes.ToLower(bytes.Trim(data, " \"")))
	b.raw = dataStr
	for _, w := range negativeWords {
		if w == dataStr {
			b.bool = false
			return nil
		}
	}

	if val, err := strconv.Atoi(dataStr); err == nil {
		b.bool = val > 0
		return nil
	}

	b.bool = true
	return nil
}

var trueBytes = bytes.NewBufferString("true").Bytes()
var falseBytes = bytes.NewBufferString("false").Bytes()

func (t LooseBool) MarshalJSON() ([]byte, error) {
	if t.bool {
		return trueBytes, nil
	} else {
		return falseBytes, nil
	}
}
