package golik

import (
	"reflect"
	ior "github.com/ioswarm/goutils/reflect"
)

func CallLifeCycle(c CloveRef, minion interface{}, methodName string) {
	if minion != nil {
		mvalue := ior.ToPtrValue(reflect.ValueOf(minion))
		if methodValue := mvalue.MethodByName(methodName); methodValue.IsValid() {
			methodType := methodValue.Type()
			switch methodType.NumIn() {
			case 0:
				methodValue.Call(nil)
			case 1:
				ctype := reflect.TypeOf(c)
				if ctype.Implements(methodType.In(0)) {
					methodValue.Call([]reflect.Value{reflect.ValueOf(c)})
				}
			}
		}
	}
}
