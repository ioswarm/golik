package golik

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

type ConvertRule interface {
	Check(reflect.Type) bool
	Decode(Converter, reflect.Value) (interface{}, error)
	Encode(Converter, interface{}, reflect.Value) error
}

type Converter interface {
	InterpretJson(bool) Converter
	NameMapping(string, string) Converter
	Capitalize(bool) Converter

	IsValid(interface{}) bool

	FieldNames(reflect.Type) map[string]string
	FieldNameMapping(interface{}) (map[string]string, error)

	AddRule(...ConvertRule) Converter
	Rules() []ConvertRule
	Decode(reflect.Value) (interface{}, error)
	Encode(interface{}, reflect.Value) error

	ToMap(interface{}) (map[string]interface{}, error)
	FromMap(map[string]interface{}, interface{}) error
}

type converter struct {
	interpretJson    bool
	fieldNameMapping map[string]string
	capitalize       bool

	rules []ConvertRule

	mutex sync.Mutex
}

func NewConverter() Converter {
	return &converter{
		interpretJson:    true,
		fieldNameMapping: make(map[string]string),
		rules: []ConvertRule{
			PtrRule(),
			TimestampRule(),
			StructRule(),
			SliceRule(),
			StringKeyMapRule(),
			MapRule(),
			StringRule(),
			BoolRule(),
			IntRule(),
			Int8Rule(),
			Int16Rule(),
			Int32Rule(),
			Int64Rule(),
			UintRule(),
			Uint8Rule(),
			Uint16Rule(),
			Uint32Rule(),
			Uint64Rule(),
			Float32Rule(),
			Float64Rule(),
		},
	}
}

func (conv *converter) validate(i interface{}) error {
	ivalue := reflect.ValueOf(i)
	if i == nil || ivalue.IsZero() || ivalue.IsNil() {
		return errors.New("Given data is nil or zero")
	}
	itype := reflect.TypeOf(i)
	if itype.Kind() != reflect.Ptr {
		return errors.New("Given data is not a pointer")
	}
	if itype.Elem().Kind() != reflect.Struct {
		return errors.New("Given data is not a pointer of struct")
	}
	return nil
}

func (conv *converter) IsValid(i interface{}) bool {
	if err := conv.validate(i); err != nil {
		return false
	}
	return true
}

func (conv *converter) FieldNames(itype reflect.Type) map[string]string {
	result := make(map[string]string)
	for i := 0; i < itype.NumField(); i++ {
		fld := itype.Field(i)
		uname := []rune(fld.Name)
		if uname[0] >= 65 && uname[0] <= 90 {
			if !conv.capitalize {
				uname[0] = uname[0] + 32
			}
			name := string(uname)
			if jsontag, ok := fld.Tag.Lookup("json"); ok && conv.interpretJson {
				name = strings.SplitN(jsontag, ",", 2)[0]
			}
			if mname, ok := conv.fieldNameMapping[name]; ok {
				name = mname
			}
			result[name] = fld.Name
		}
	}
	return result
}

func (conv *converter) FieldNameMapping(i interface{}) (map[string]string, error) {
	result := make(map[string]string)
	if err := conv.validate(i); err != nil {
		return result, err
	}
	result = conv.FieldNames(reflect.TypeOf(i))
	return result, nil
}

func (conv *converter) InterpretJson(b bool) Converter {
	conv.mutex.Lock()
	defer conv.mutex.Unlock()

	conv.interpretJson = b
	return conv
}

func (conv *converter) NameMapping(from string, to string) Converter {
	conv.mutex.Lock()
	defer conv.mutex.Unlock()

	conv.fieldNameMapping[from] = to
	return conv
}

func (conv *converter) Capitalize(b bool) Converter {
	conv.mutex.Lock()
	defer conv.mutex.Unlock()

	conv.capitalize = b
	return conv
}

func (conv *converter) AddRule(rule ...ConvertRule) Converter {
	conv.mutex.Lock()
	defer conv.mutex.Unlock()

	conv.rules = append(rule, conv.rules...)
	return conv
}

func (conv *converter) Rules() []ConvertRule {
	return conv.rules
}

func (conv *converter) Decode(value reflect.Value) (interface{}, error) {
	vtype := value.Type()
	for _, rule := range conv.Rules() {
		if rule.Check(vtype) {
			return rule.Decode(conv, value)
		}
	}
	return nil, fmt.Errorf("Can not convert %t to interface{}", vtype)
}

func (conv *converter) Encode(i interface{}, value reflect.Value) error {
	/*if i == nil {
		return errors.New("Given interface{} is null")
	}*/
	tpe := value.Type()
	for _, rule := range conv.Rules() {
		if rule.Check(tpe) {
			return rule.Encode(conv, i, value)
		}
	}
	return fmt.Errorf("Can not encode %T", i)
}

func (conv *converter) ToMap(i interface{}) (map[string]interface{}, error) {
	if err := conv.validate(i); err != nil {
		return nil, err
	}

	value := reflect.ValueOf(i)
	resi, err := conv.Decode(value)
	if err != nil {
		return nil, err
	}
	if result, ok := resi.(map[string]interface{}); ok {
		return result, nil
	}
	return nil, fmt.Errorf("Could not cast %T to map[string]interface{}", resi)
}

func (conv *converter) FromMap(smap map[string]interface{}, i interface{}) error {
	itype := reflect.TypeOf(i)
	if itype.Kind() != reflect.Ptr {
		return errors.New("Given interface is not a pointer")
	}
	if itype.Elem().Kind() != reflect.Struct {
		return errors.New("Given interface must be a pointer of struct")
	}
	ptrvalue := reflect.New(itype)
	if i != nil {
		ptrvalue = reflect.ValueOf(i)
	}

	return conv.Encode(smap, ptrvalue)
}
