package test

import (
	"encoding/json"
	"testing"

	"github.com/c1emon/gcommon/data"
)

type dto struct {
	Name   string
	Active data.LooseBool
}

func Test_bool(t *testing.T) {
	def := data.LooseBool{}
	t.Logf("def val is %t", def.ToBool())

	d := &dto{
		Name:   "123",
		Active: data.NewLooseBool(false),
	}

	b, err := json.Marshal(d)
	if err != nil {
		t.Error(err)
		return
	}

	t.Log(string(b))

	input := `{"Name":"123","Active":"good"}`

	var h dto
	err = json.Unmarshal([]byte(input), &h)
	if err != nil {
		t.Error(err)
		return
	}
	t.Logf("val is %t", h.Active.ToBool())
}
