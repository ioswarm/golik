package golik

import (
	"reflect"
	"sort"

	"github.com/ioswarm/golik/utils"
)

func CallLifeCycle(obj interface{}, methodName string, ctx CloveContext) {
	if obj != nil {
		objValue := utils.ToPtrValue(reflect.ValueOf(obj))
		if methodValue := objValue.MethodByName(methodName); methodValue.IsValid() {
			methodType := methodValue.Type()
			switch methodType.NumIn() {
			case 0:
				methodValue.Call(nil)
			case 1:
				ctype := reflect.TypeOf(ctx)
				if ctype.Implements(methodType.In(0)) {
					methodValue.Call([]reflect.Value{reflect.ValueOf(ctx)})
				}
			}
		}
	}
}

func CallMethod(obj interface{}, ctx CloveContext, input interface{}) (interface{}, bool) {
	if obj != nil {
		objValue := utils.ToPtrValue(reflect.ValueOf(obj))
		inputValue := reflect.ValueOf(input)
		ctxValue := reflect.ValueOf(ctx)
		meths := utils.FindMethodsOf(objValue, inputValue.Type())
		meths = append(meths, utils.FindMethodsOf(objValue, inputValue.Type(), ctxValue.Type())...)
		if len(meths) > 0 {
			methodCategory := func(meth reflect.Value) int {
				mod := 1
				methType := meth.Type()
				if methType.NumIn() == 2 && ctxValue.Type().Implements(methType.In(1)) {
					mod = 0
				}
				if methType.NumOut() == 2 && utils.IsErrorType(methType.Out(1)) {
					return 0+mod
				} else if methType.NumOut() == 2 && utils.IsErrorType(methType.Out(0)) {
					return 2+mod
				} else if methType.NumOut() == 1 && utils.IsErrorType(methType.Out(0)) {
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

			meth := meths[0]
			switch methodCategory(meth) {
			case 0:
				// TODO debug-internal ctx.Debug("Method-state-0")
				result := meth.Call([]reflect.Value{inputValue, ctxValue})
				if err, ok := result[1].Interface().(error); ok {
					return err, true
				}
				return result[0].Interface(), true
			case 1:
				// TODO debug-internal ctx.Debug("Method-state-1")
				result := meth.Call([]reflect.Value{inputValue})
				if err, ok := result[1].Interface().(error); ok {
					return err, true
				}
				return result[0].Interface(), true
			case 2:
				// TODO debug-internal ctx.Debug("Method-state-2")
				result := meth.Call([]reflect.Value{inputValue, ctxValue})
				if err, ok := result[0].Interface().(error); ok {
					return err, true
				}
				return result[1].Interface(), true
			case 3:
				// TODO debug-internal ctx.Debug("Method-state-3")
				result := meth.Call([]reflect.Value{inputValue})
				if err, ok := result[0].Interface().(error); ok {
					return err, true
				}
				return result[1].Interface(), true
			case 4:
				// TODO debug-internal ctx.Debug("Method-state-4")
				result := meth.Call([]reflect.Value{inputValue, ctxValue})
				if err, ok := result[0].Interface().(error); ok {
					return err, true
				}
				return nil, true
			case 5:
				// TODO debug-internal ctx.Debug("Method-state-5")
				result := meth.Call([]reflect.Value{inputValue})
				if err, ok := result[0].Interface().(error); ok {
					return err, true
				}
				return nil, true
			case 6:
				// TODO debug-internal ctx.Debug("Method-state-6")
				result := meth.Call([]reflect.Value{inputValue, ctxValue})
				return result[0].Interface(), true
			case 7:
				// TODO debug-internal ctx.Debug("Method-state-7")
				result := meth.Call([]reflect.Value{inputValue})
				return result[0].Interface(), true
			case 8:
				// TODO debug-internal ctx.Debug("Method-state-8")
				meth.Call([]reflect.Value{inputValue, ctxValue})
				return nil, true
			case 9:
				// TODO debug-internal ctx.Debug("Method-state-9")
				meth.Call([]reflect.Value{inputValue})
				return nil, true
			default:
				// TODO debug-internal ctx.Debug("Method-state-? not match")
				// TODO throw warning ... method found but too many result-types
				return nil, false
			}
		}
		return nil, false
	}
	return nil, false
}
