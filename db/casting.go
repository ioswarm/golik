package db

import (
	"fmt"
	"reflect"
)

func isValidIntKind(kind reflect.Kind) bool {
	return kind == reflect.Int ||
		kind == reflect.Int8 ||
		kind == reflect.Int16 ||
		kind == reflect.Int32 ||
		kind == reflect.Int64 ||
		kind == reflect.Uint ||
		kind == reflect.Uint8 ||
		kind == reflect.Uint16 ||
		kind == reflect.Uint32 ||
		kind == reflect.Uint64
}

func intBase(a interface{}) int64 {
	switch a.(type) {
	case int:
		return int64(a.(int))
	case int8:
		return int64(a.(int8))
	case int16:
		return int64(a.(int16))
	case int32:
		return int64(a.(int32))
	case int64:
		return a.(int64)
	case uint:
		return int64(a.(uint))
	case uint8:
		return int64(a.(uint8))
	case uint16:
		return int64(a.(uint16))
	case uint32:
		return int64(a.(uint32))
	case uint64:
		return int64(a.(uint64))
	}
	return int64(0)
}

func uintBase(a interface{}) uint64 {
	switch a.(type) {
	case int:
		return uint64(a.(int))
	case int8:
		return uint64(a.(int8))
	case int16:
		return uint64(a.(int16))
	case int32:
		return uint64(a.(int32))
	case int64:
		return uint64(a.(int64))
	case uint:
		return uint64(a.(uint))
	case uint8:
		return uint64(a.(uint8))
	case uint16:
		return uint64(a.(uint16))
	case uint32:
		return uint64(a.(uint32))
	case uint64:
		return a.(uint64)
	}
	return uint64(0)
}

func CastI(a interface{}, as reflect.Kind) (interface{}, error) {
	atype := reflect.TypeOf(a)
	if !isValidIntKind(atype.Kind()) {
		return nil, fmt.Errorf("Unsupported source type %v", atype)
	}
	if !isValidIntKind(as) {
		return nil, fmt.Errorf("Can not cast to %v", as)
	}
	switch as {
	case reflect.Int:
		return int(intBase(a)), nil
	case reflect.Int8:
		return int8(intBase(a)), nil
	case reflect.Int16:
		return int16(intBase(a)), nil
	case reflect.Int32:
		return int32(intBase(a)), nil
	case reflect.Uint:
		return uint(uintBase(a)), nil
	case reflect.Uint8:
		return uint8(uintBase(a)), nil
	case reflect.Uint16:
		return uint16(uintBase(a)), nil
	case reflect.Uint32:
		return uint32(uintBase(a)), nil
	case reflect.Uint64:
		return uintBase(a), nil
	default:
		return intBase(a), nil
	}
}