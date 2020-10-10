package golik

import (
	"reflect"
)

var (
	errorType   = reflect.TypeOf((*error)(nil)).Elem()
	contextType = reflect.TypeOf((*CloveContext)(nil)).Elem()
	messageType = reflect.TypeOf((*Message)(nil)).Elem()
)

func IsErrorType(gtype reflect.Type) bool {
	return gtype.Implements(errorType)
}

func IsContextType(gtype reflect.Type) bool {
	return gtype.Implements(contextType)
}

func IsMessageType(gtype reflect.Type) bool {
	return gtype.Implements(messageType)
}

func checkImplements(atype reflect.Type, btype reflect.Type) bool {
	if btype.Kind() == reflect.Interface {
		return atype.Implements(btype)
	} else if atype.Kind() == reflect.Interface {
		return btype.Implements(atype)
	}
	return false
}

func CompareType(atype reflect.Type, btype reflect.Type) bool {
	return atype == btype || checkImplements(atype, btype)
}

func toPtrValue(gvalue reflect.Value) reflect.Value {
	if gvalue.Kind() != reflect.Ptr {
		pto := reflect.PtrTo(gvalue.Type())
		ptr := reflect.New(pto.Elem())
		ptr.Elem().Set(gvalue)
		return ptr
	}
	return gvalue
}

func checkParams(ftype reflect.Type, params ...reflect.Type) bool {
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

func findMethodsOf(gvalue reflect.Value, params ...reflect.Type) []reflect.Value {
	result := make([]reflect.Value, 0)
	ptrvalue := toPtrValue(gvalue)
	if ptrvalue.Kind() == reflect.Ptr && ptrvalue.Elem().Kind() == reflect.Struct {
		for i := 0; i < ptrvalue.NumMethod(); i++ {
			mvalue := ptrvalue.Method(i)
			if checkParams(mvalue.Type(), params...) {
				result = append(result, mvalue)
			}
		}
	}
	return result
}

func findMethodByName(gvalue reflect.Value, methodName string, params ...reflect.Type) (reflect.Value, bool) {
	ptrvalue := toPtrValue(gvalue)
	if ptrvalue.Kind() == reflect.Ptr && ptrvalue.Elem().Kind() == reflect.Struct {
		if mv := ptrvalue.MethodByName(methodName); mv.IsValid() {
			if checkParams(mv.Type(), params...) {
				return mv, true
			}
		}
	}
	return reflect.ValueOf(nil), false
}
