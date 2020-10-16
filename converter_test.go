package golik

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

type ComplexTypeA struct {
	Id string
	StringValue string
	Struct ComplexTypeB
	SliceNative []int
	SliceStruct []ComplexTypeB
	SlicePtr []*ComplexTypeB
	SimpleMap map[int]string
}

func newComplexTypeA() *ComplexTypeA {
	return &ComplexTypeA{
		Id: "this_is_a_id", 
		StringValue: "This is a stringValue",
		Struct: newComplexTypeB(),
		SliceNative: []int{2, 4, 6, 8, 10},
		SliceStruct: []ComplexTypeB{
			{1.234, 4.321, 20200920, 16},
		},
		SlicePtr: []*ComplexTypeB{
			{1.234, 4.321, 20200920, 16},
		},
		SimpleMap: map[int]string{
			1: "one",
			2: "two",
			3: "three",
		},
	}
}

type ComplexTypeB struct {
	Float32Value float32
	Float64Value float64
	IntValue int
	Int8Value int8
}

func newComplexTypeB() ComplexTypeB {
	return ComplexTypeB{
		Float32Value: 3.1415,
		Float64Value: 123424.32483,
		Int8Value: 16,
		IntValue: 20200920,
	}
}

func TestIsValid(t *testing.T) {
	conv := NewConverter()
	if conv.IsValid(nil) {
		t.Errorf("If given data is nil, IsValid must return %v", false)
	}
	var v int 
	if conv.IsValid(v) {
		t.Errorf("If given data is not a pointer, IsValid must return %v", false)
	}
	if conv.IsValid(&v) {
		t.Errorf("If given data is not a struct, IsValid must return %v", false)
	}
}

func TestConvertComplexTypeAToMap(t *testing.T) {
	conv := NewConverter()
	ct := newComplexTypeA()
	m, err := conv.ToMap(ct)
	if err != nil {
		t.Errorf("Conversion %T to map, must not return error %v", ct, err)
	}
	if _, ok := m["id"]; !ok {
		t.Errorf("map[id] must exists")
	}
	t.Logf("Map is %v", m)
}

func TestConvertComplexTypeAFromMap(t *testing.T) {
	assert := assert.New(t)

	mvalue := make(map[string]interface{})
	mvalue["id"] = "4711-0815"
	mvalue["stringValue"] = "Hello there!"
	mvalue["struct"] = map[string]interface{} {
		"float32Value": float32(123.456),
	}
	mvalue["sliceNative"] = []interface{}{1,3,5,7,9}
	mvalue["sliceStruct"] = []interface{}{
		map[string]interface{}{
			"float32Value": float32(1.234), 
			"float64Value": 4.321,
			"intValue": 20200920,
			"int8Value": int8(8),
		},
	}
	mvalue["slicePtr"] = []interface{}{
		map[string]interface{}{
			"float32Value": float32(1.234), 
			"float64Value": 4.321,
			"intValue": 20200920,
			"int8Value": int8(8),
		},
	}
	mvalue["simpleMap"] = map[interface{}]interface{}{
		10: "ten",
		20: "twenty",
		40: "fourty",
	}

	conv := NewConverter()
	var c ComplexTypeA

	if err := conv.FromMap(mvalue, c); err == nil {
		t.Error("Converter.FromMap with no given pointer must return an error")		
		return
	}

	if err := conv.FromMap(mvalue, &c); err != nil {
		t.Errorf("Converter.FromMap must not return an error, when called with map and pointer - error was: %v", err)
		return 
	}

	assert.Equal(mvalue["id"], c.Id)
	assert.Equal(mvalue["stringValue"], c.StringValue)
	assert.Equal(float32(123.456), c.Struct.Float32Value)

	assert.Equal(float32(1.234), c.SlicePtr[0].Float32Value)

	t.Logf("ComplexTypeA is: %v", c)
}