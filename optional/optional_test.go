package optional

import (
	"encoding/json"
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

func TestOptional_NewFromPtr(t *testing.T) {
	if NewFromPtr[int](nil).Present() {
		t.Fatal("NewFromPtr(nil): want absent")
	}
	n := 42
	o := NewFromPtr(&n)
	v, err := o.Get()
	if err != nil || v != 42 {
		t.Fatalf("Get: (%d, %v), want (42, nil)", v, err)
	}
}

func TestOptional_OrElse(t *testing.T) {
	if got := (Optional[int]{}).OrElse(3); got != 3 {
		t.Fatalf("OrElse absent: got %d, want 3", got)
	}
	if got := New(5).OrElse(3); got != 5 {
		t.Fatalf("OrElse present: got %d, want 5", got)
	}
}

func TestOptional_If(t *testing.T) {
	var got int
	var o Optional[int]
	o.If(func(v int) { got = v })
	if got != 0 {
		t.Fatalf("If absent: got %d, want 0", got)
	}
	New(9).If(func(v int) { got = v })
	if got != 9 {
		t.Fatalf("If present: got %d, want 9", got)
	}
}

func TestOptional_ToPtr_copy(t *testing.T) {
	o := New(1)
	p := o.ToPtr()
	if p == nil {
		t.Fatal("ToPtr: nil")
	}
	*p = 2
	v, _ := o.Get()
	if v != 1 {
		t.Fatalf("mutating ToPtr result changed inner value: got %d, want 1", v)
	}
	if (Optional[int]{}).ToPtr() != nil {
		t.Fatal("ToPtr absent: want nil")
	}
}

func TestOptional_Clear(t *testing.T) {
	o := New(1)
	o.Clear()
	if o.Present() {
		t.Fatal("after Clear: want absent")
	}
}

func TestOptional_MarshalJSON(t *testing.T) {
	b, err := json.Marshal(New(10))
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "10" {
		t.Fatalf("Marshal present: %s, want 10", b)
	}
	b, err = json.Marshal(Optional[int]{})
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "null" {
		t.Fatalf("Marshal absent: %s, want null", b)
	}
}

func TestOptional_UnmarshalJSON_nullVariants(t *testing.T) {
	for _, raw := range []string{"null", " null ", "\tnull\n"} {
		var o Optional[int]
		if err := json.Unmarshal([]byte(raw), &o); err != nil {
			t.Fatalf("Unmarshal %q: %v", raw, err)
		}
		if o.Present() {
			t.Fatalf("Unmarshal %q: want absent", raw)
		}
	}
}

func TestOptional_UnmarshalJSON_value(t *testing.T) {
	var o Optional[int]
	if err := json.Unmarshal([]byte("7"), &o); err != nil {
		t.Fatal(err)
	}
	v, err := o.Get()
	if err != nil || v != 7 {
		t.Fatalf("Get: (%d, %v), want (7, nil)", v, err)
	}
}

func TestOptional_JSON_roundTrip(t *testing.T) {
	type row struct {
		X Optional[string] `json:"x"`
	}
	in := row{X: New("hi")}
	b, err := json.Marshal(in)
	if err != nil {
		t.Fatal(err)
	}
	var out row
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if got, _ := out.X.Get(); got != "hi" {
		t.Fatalf("round trip: got %q, want hi", got)
	}
}

func TestOptional_MustGet_panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("MustGet absent: want panic")
		}
	}()
	(Optional[int]{}).MustGet()
}
