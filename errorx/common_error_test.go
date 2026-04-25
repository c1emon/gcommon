package errorx

import (
	"errors"
	"fmt"
	"testing"
)

func TestCommonError_Is_sameCode(t *testing.T) {
	a := NewCommonError(42, "a")
	b := NewCommonError(42, "b")
	if !errors.Is(a, b) {
		t.Fatal("errors.Is: expect match on same business code")
	}
}

func TestCommonError_Is_differentCode(t *testing.T) {
	a := NewCommonError(42, "a")
	c := NewCommonError(43, "c")
	if errors.Is(a, c) {
		t.Fatal("errors.Is: expect no match on different code")
	}
}

func TestCommonError_Is_wrapped(t *testing.T) {
	a := NewCommonError(42, "a")
	b := NewCommonError(42, "b")
	w := fmt.Errorf("outer: %w", a)
	if !errors.Is(w, b) {
		t.Fatal("errors.Is: expect match through unwrap chain")
	}
}

func TestCommonError_Is_httpErrorTarget(t *testing.T) {
	inner := NewCommonError(1002, "illegal param")
	if !errors.Is(inner, ErrHttpIllegalParam) {
		t.Fatal("errors.Is: *CommonError should match *HttpError with same code")
	}
}
