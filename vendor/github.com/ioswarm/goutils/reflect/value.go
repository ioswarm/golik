package reflect

import (
	"reflect"
)

func IsErrorValue(gvalue reflect.Value) bool {
	return IsErrorType(gvalue.Type())
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

func FindProceduresOf(gvalue reflect.Value, params ...reflect.Type) []reflect.Value {
	result := make([]reflect.Value, 0)
	for _, tfunc := range FindMethodsOf(gvalue, params...) {
		if tfunc.Type().NumOut() == 0 {
			result = append(result, tfunc)
		}
	}
	return result
}

func FindFunctionsOf(gvalue reflect.Value, params ...reflect.Type) []reflect.Value {
	result := make([]reflect.Value, 0)
	for _, tfunc := range FindMethodsOf(gvalue, params...) {
		if tfunc.Type().NumOut() > 0 {
			result = append(result, tfunc)
		}
	}
	return result
}