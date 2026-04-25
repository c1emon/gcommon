package vo

import (
	"encoding/json"
	"testing"
)

func TestResult_any_JSON(t *testing.T) {
	var r Result[any]
	err := json.Unmarshal([]byte(`{"code":0,"msg":"ok","ts":1}`), &r)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWrapPagination(t *testing.T) {
	inner := NewResultOK([]string{"a"})
	p := WrapPagination(*inner)
	if len(p.Data.MustGet()) != 1 || p.Data.MustGet()[0] != "a" {
		t.Fatalf("data: %+v", p.Data)
	}
}
