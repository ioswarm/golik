package golik

import (
	"reflect"
	"sort"
	ior "github.com/ioswarm/goutils/reflect"
)

func callLifeCycle(c CloveRef, minion interface{}, methodName string) {
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

func callMinionLogic(c CloveRef, minion interface{}, value interface{}) (interface{}, error) {
	if minion != nil {
		mvalue := ior.ToPtrValue(reflect.ValueOf(minion))
		vvalue := reflect.ValueOf(value)
		cvalue := reflect.ValueOf(c)

		meths := ior.FindMethodsOf(mvalue, vvalue.Type())
		meths = append(meths, ior.FindMethodsOf(mvalue, vvalue.Type(), cvalue.Type())...)
		if len(meths) > 0 { 
			methodCategory := func(meth reflect.Value) int {
				mod := 1
				methType := meth.Type()
				if methType.NumIn() == 2 && cvalue.Type().Implements(methType.In(1)) {
					mod = 0
				}
				if methType.NumOut() == 2 && ior.IsErrorType(methType.Out(1)) {
					return 0+mod
				} else if methType.NumOut() == 2 && ior.IsErrorType(methType.Out(0)) {
					return 2+mod
				} else if methType.NumOut() == 1 && !ior.IsErrorType(methType.Out(0)) {
					return 4+mod
				} else if methType.NumOut() == 1 { 
					return 6+mod
				} else if methType.NumOut() == 0 { 
					return 8+mod
				}
				return 999
			}

			sort.Slice(meths, func(a, b int) bool {
				return methodCategory(meths[a]) < methodCategory(meths[b])
			})

			// TODO ... debug info
			meth := meths[0]
			switch methodCategory(meth) {
			case 0:
				result := meth.Call([]reflect.Value{vvalue, cvalue})
				if err, ok := result[1].Interface().(error); ok {
					return result[0].Interface(), err
				}
				return result[0].Interface(), nil
			case 1:
				result := meth.Call([]reflect.Value{vvalue})
				if err, ok := result[1].Interface().(error); ok {
					return result[0].Interface(), err
				}
				return result[0].Interface(), nil
			case 2:
				result := meth.Call([]reflect.Value{vvalue, cvalue})
				if err, ok := result[0].Interface().(error); ok {
					return result[1].Interface(), err
				}
				return result[1].Interface(), nil
			case 3:
				result := meth.Call([]reflect.Value{vvalue})
				if err, ok := result[0].Interface().(error); ok {
					return result[1].Interface(), err
				}
				return result[1].Interface(), nil
			case 4:
				result := meth.Call([]reflect.Value{vvalue, cvalue})
				return result[0].Interface(), nil
			case 5:
				result := meth.Call([]reflect.Value{vvalue})
				return result[0].Interface(), nil
			case 6:
				result := meth.Call([]reflect.Value{vvalue, cvalue})
				if err, ok := result[0].Interface().(error); ok {
					return nil, err
				}
				return nil, nil
			case 7:
				result := meth.Call([]reflect.Value{vvalue})
				if err, ok := result[0].Interface().(error); ok {
					return nil, err
				}
				return nil, nil
			case 8:
				meth.Call([]reflect.Value{vvalue, cvalue})
				return nil, nil
			case 9:
				meth.Call([]reflect.Value{vvalue})
				return nil, nil
			default:
				// TODO throw warning ... method found but too many result-types
			}
		}
	}
	return nil, nil
}

func callMinionRoutes(minion interface{}) []*Route {
	tr := reflect.TypeOf((*Route)(nil))

	result := make([]*Route, 0)
	mvalue := reflect.ValueOf(minion)
	for i:=0;i<mvalue.NumMethod();i++ {
		methvalue := mvalue.Method(i)
		methtype := methvalue.Type()
		if methtype.NumIn() == 0 && methtype.NumOut() == 1 && methtype.Out(0) == tr {
			methresult := methvalue.Call(nil)
			if len(methresult) == 1 {
				result = append(result, methresult[0].Interface().(*Route))
			}
		}
	}
	return result
}

func callMinionCloveRoutes(minion interface{}) []CloveRoute {
	tr := reflect.TypeOf((CloveRoute)(nil))

	result := make([]CloveRoute, 0)
	mvalue := reflect.ValueOf(minion)
	for i:=0;i<mvalue.NumMethod();i++ {
		methvalue := mvalue.Method(i)
		methtype := methvalue.Type()
		if methtype.NumIn() == 0 && methtype.NumOut() == 1 && methtype.Out(0) == tr {
			methresult := methvalue.Call(nil)
			if len(methresult) == 1 {
				result = append(result, methresult[0].Interface().(CloveRoute))
			}
		}
	}
	return result
}