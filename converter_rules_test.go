package golik

import (
	"reflect"
	"testing"
	"time"
)

func TestTimeRule(t *testing.T) {
	rule := TimeToStringRule()
	conv := NewConverter().AddRule(rule)

	tn := time.Date(2008, 2, 10, 18, 5, 15,224000, time.UTC)

	if !rule.Check(reflect.TypeOf(tn)) {
		t.Error("Datatype check of time-type must return true")
	}

	dec, err := rule.Decode(conv, reflect.ValueOf(tn))
	if err != nil {
		t.Errorf("Decode time should not return error: %v", err)
	}

	res, ok := dec.(string)
	if !ok {
		t.Errorf("Decode time must return value of string not %T", res)
	}

	var tx time.Time
	txptr := reflect.ValueOf(&tx)
	err = rule.Encode(conv, res, txptr.Elem())
	if err != nil {
		t.Errorf("Encode string to time.Time must not return an error: %v", err)
	}

	type C struct {
		Time time.Time
	}
	var c C
	cv := reflect.ValueOf(&c).Elem()
	rule.Encode(conv, res, cv.FieldByName("Time"))
	if err != nil {
		t.Errorf("Encode string to time.Time struct field must not return an error: %v", err)
	}
}