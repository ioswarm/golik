package reflect

import (
	"reflect"
)

func IsErrorType(gtype reflect.Type) bool {
	return gtype.Implements(errorType)
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