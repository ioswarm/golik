package utils

import (
	"reflect"
)

var (
	errorType = reflect.TypeOf((*error)(nil)).Elem()
)

func IsErrorType(gtype reflect.Type) bool {
	return gtype.Implements(errorType)
}

func IsErrorValue(gvalue reflect.Value) bool {
	return IsErrorType(gvalue.Type())
}

func CheckImplements(atype reflect.Type, btype reflect.Type) bool {
	if btype.Kind() == reflect.Interface {
		return atype.Implements(btype)
	} else if atype.Kind() == reflect.Interface {
		return btype.Implements(atype)
	}
	return false
}

func CompareType(atype reflect.Type, btype reflect.Type) bool {
	return atype == btype || CheckImplements(atype, btype)
}

func ToPtrValue(gvalue reflect.Value) reflect.Value {
	if gvalue.Kind() != reflect.Ptr {
		pto := reflect.PtrTo(gvalue.Type())
		ptr := reflect.New(pto.Elem())
		ptr.Elem().Set(gvalue)
		return ptr
	}
	return gvalue
}

func FindMethodsOf(gvalue reflect.Value, params ...reflect.Type) []reflect.Value {
	result := make([]reflect.Value, 0)
	ptrvalue := ToPtrValue(gvalue)

	checkParams := func(ftype reflect.Type) bool {
		if ftype.NumIn() == len(params) {
			for i, ttype := range params {
				if !CompareType(ftype.In(i), ttype) {
					return false
				}
			}
			return true
		}
		return false
	}

	for i := 0; i < ptrvalue.NumMethod(); i++ {
		mvalue := ptrvalue.Method(i)
		if checkParams(mvalue.Type()) {
			result = append(result, mvalue)
		}
	}
	return result
}


