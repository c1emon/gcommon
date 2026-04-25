package optional

import (
	"errors"
	"testing"
)

func TestOptional_Get_ErrNotPresent(t *testing.T) {
	var o Optional[int]
	_, err := o.Get()
	if !errors.Is(err, ErrNotPresent) {
		t.Fatalf("Get: got %v, want ErrNotPresent", err)
	}
}

func TestOptional_Get_ok(t *testing.T) {
	o := New(7)
	v, err := o.Get()
	if err != nil || v != 7 {
		t.Fatalf("Get: (%d, %v), want (7, nil)", v, err)
	}
}
