package golik

import (
	"context"
	"fmt"
	"reflect"
)

type BehaviorCategory int

const (
	UNKNOWN BehaviorCategory = iota

	NONEWITHNONE
	NONEWITHRESULT
	NONEWITHRESULTERROR

	CONTEXTWITHNONE
	CONTEXTWITHRESULT
	CONTEXTWITHRESULTERROR

	MSGWITHNONE

	DATAWITHNONE
	DATAWITHRESULT
	DATAWITHRESULTERROR

	CONTEXTMSGWITHNONE

	CONTEXTDATAWITHNONE
	CONTEXTDATAWITHRESULT
	CONTEXTDATAWITHRESULTERROR

	PTR
	STRUCT
)

func behaviorCategory(ftype reflect.Type) BehaviorCategory {
	//ftype := reflect.TypeOf(f)

	if ftype.Kind() == reflect.Func {
		if ftype.NumIn() == 0 && ftype.NumOut() == 0 {
			return NONEWITHNONE
		}
		if ftype.NumIn() == 0 && ftype.NumOut() == 1 {
			return NONEWITHRESULT
		}
		if ftype.NumIn() == 0 && ftype.NumOut() == 2 && IsErrorType(ftype.Out(1)) {
			return NONEWITHRESULTERROR
		}

		if ftype.NumIn() == 1 && IsContextType(ftype.In(0)) && ftype.NumOut() == 0 {
			return CONTEXTWITHNONE
		}
		if ftype.NumIn() == 1 && IsContextType(ftype.In(0)) && ftype.NumOut() == 1 {
			return CONTEXTWITHRESULT
		}
		if ftype.NumIn() == 1 && IsContextType(ftype.In(0)) && ftype.NumOut() == 2 && IsErrorType(ftype.Out(1)) {
			return CONTEXTWITHRESULTERROR
		}

		if ftype.NumIn() == 1 && IsMessageType(ftype.In(0)) && ftype.NumOut() == 0 {
			return MSGWITHNONE
		}

		if ftype.NumIn() == 1 && ftype.NumOut() == 0 {
			return DATAWITHNONE
		}
		if ftype.NumIn() == 1 && ftype.NumOut() == 1 {
			return DATAWITHRESULT
		}
		if ftype.NumIn() == 1 && ftype.NumOut() == 2 && IsErrorType(ftype.Out(1)) {
			return DATAWITHRESULTERROR
		}

		if ftype.NumIn() == 2 && IsContextType(ftype.In(0)) && IsMessageType(ftype.In(1)) && ftype.NumOut() == 0 {
			return CONTEXTMSGWITHNONE
		}

		if ftype.NumIn() == 2 && IsContextType(ftype.In(0)) && ftype.NumOut() == 0 {
			return CONTEXTDATAWITHNONE
		}
		if ftype.NumIn() == 2 && IsContextType(ftype.In(0)) && ftype.NumOut() == 1 {
			return CONTEXTDATAWITHRESULT
		}
		if ftype.NumIn() == 2 && IsContextType(ftype.In(0)) && ftype.NumOut() == 2 && IsErrorType(ftype.Out(1)) {
			return CONTEXTDATAWITHRESULTERROR
		}
	}

	if ftype.Kind() == reflect.Ptr && ftype.Elem().Kind() == reflect.Struct {
		return PTR
	}

	if ftype.Kind() == reflect.Struct {
		return STRUCT
	}

	return UNKNOWN
}

func checkBehavior(b interface{}) error {
	if b == nil {
		return Errorln("Behavior is not defined")
	}
	checkResult := func(ftype reflect.Type) bool {
		switch ftype.NumOut() {
		case 0, 1:
			return true
		case 2:
			if IsErrorType(ftype.Out(1)) {
				return true
			}
		}
		return false
	}

	btype := reflect.TypeOf(b)
	switch btype.Kind() {
	case reflect.Func:
		switch btype.NumIn() {
		case 1:
			if checkResult(btype) {
				return nil
			}
		case 2:
			if checkResult(btype) {
				if IsContextType(btype.In(0)) {
					return nil
				}
			}
		}
		return Errorf("Unsupported func %T", b)
	case reflect.Struct:
		return nil
	case reflect.Ptr:
		if btype.Elem().Kind() != reflect.Struct {
			return Errorf("Only pointer of struct are supported, found %T", b)
		}
		return nil
	}
	return Errorf("Unsupported type %T to handle Behavior", b)
}

func checkLifecycleFunc(b interface{}) error {
	if b == nil {
		return nil
	}
	btype := reflect.TypeOf(b)
	if btype.Kind() == reflect.Func {
		if btype.NumOut() == 0 || (btype.NumOut() == 1 && IsErrorType(btype.Out(0))) {
			if btype.NumIn() == 0 || (btype.NumIn() == 1 && IsContextType(btype.In(0))) {
				return nil
			}
		}
	}
	return Errorf("Unsupported type %T to handle lifecycle calls", b)
}

func executeLifecycle(ctx CloveContext, f interface{}) error {
	if f == nil {
		return Errorf("Lifecycle function is not defined")
	}
	if err := checkLifecycleFunc(f); err != nil {
		return err
	}

	ftype := reflect.TypeOf(f)
	fvalue := reflect.ValueOf(f)

	switch ftype.NumIn() {
	case 0:
		result := fvalue.Call(nil)
		if ftype.NumOut() == 1 && IsErrorType(ftype.Out(0)) {
			return result[0].Interface().(error)
		}
	case 1:
		ctxvalue := reflect.ValueOf(ctx)
		result := fvalue.Call([]reflect.Value{ctxvalue})
		if ftype.NumOut() == 1 && IsErrorType(ftype.Out(0)) {
			res := result[0].Interface()
			if res != nil {
				return res.(error)
			}
			return nil
		}
	}

	return nil
}

func callLifecycle(ctx CloveContext, f interface{}) error {
	c := make(chan error, 1)
	go func() { c <- executeLifecycle(ctx, f) }()
	select {
	case <-ctx.Done():
		<-c
		return ctx.Err()
	case err := <-c:
		return err
	}
}

func CallBehavior(ctx CloveContext, msg Message, f interface{}) {
	callBehaviorValue(ctx, msg, reflect.ValueOf(f))
}

func callBehaviorValue(ctx CloveContext, msg Message, fvalue reflect.Value) {
	singleResult := func(result []reflect.Value) {
		msg.Reply(result[0].Interface())
	}
	tupleResult := func(result []reflect.Value) {
		if !result[1].IsNil() {
			msg.Reply(result[1].Interface())
			return
		}
		msg.Reply(result[0].Interface())
	}

	ftype := fvalue.Type()
	switch behaviorCategory(ftype) {
	case NONEWITHNONE:
		//fmt.Println("Call none with none")
		fvalue.Call(nil)
		msg.Reply(Done())
	case NONEWITHRESULT:
		//fmt.Println("Call none with result")
		singleResult(fvalue.Call(nil))
	case NONEWITHRESULTERROR:
		//fmt.Println("Call none with result,error")
		tupleResult(fvalue.Call(nil))
	case CONTEXTWITHNONE:
		//fmt.Println("Call context with none")
		fvalue.Call([]reflect.Value{reflect.ValueOf(ctx)})
		msg.Reply(Done())
	case CONTEXTWITHRESULT:
		//fmt.Println("Call context with result")
		singleResult(fvalue.Call([]reflect.Value{reflect.ValueOf(ctx)}))
	case CONTEXTWITHRESULTERROR:
		//fmt.Println("Call context with result,error")
		tupleResult(fvalue.Call([]reflect.Value{reflect.ValueOf(ctx)}))
	case MSGWITHNONE:
		//fmt.Println("Call message with none")
		fvalue.Call([]reflect.Value{reflect.ValueOf(msg)})
	case DATAWITHNONE:
		//fmt.Println("Call data with none")
		fvalue.Call([]reflect.Value{reflect.ValueOf(msg.Content())})
		msg.Reply(Done())
	case DATAWITHRESULT:
		//fmt.Println("Call data with result")
		singleResult(fvalue.Call([]reflect.Value{reflect.ValueOf(msg.Content())}))
	case DATAWITHRESULTERROR:
		//fmt.Println("Call data with result,error")
		tupleResult(fvalue.Call([]reflect.Value{reflect.ValueOf(msg.Content())}))
	case CONTEXTMSGWITHNONE:
		//fmt.Println("Call context message with none")
		fvalue.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(msg)})
	case CONTEXTDATAWITHNONE:
		//fmt.Println("Call context data with none")
		fvalue.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(msg.Content())})
		msg.Reply(Done())
	case CONTEXTDATAWITHRESULT:
		//fmt.Println("Call context data with result")
		singleResult(fvalue.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(msg.Content())}))
	case CONTEXTDATAWITHRESULTERROR:
		//fmt.Println("Call context data with result,error")
		tupleResult(fvalue.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(msg.Content())}))
	case PTR, STRUCT:
		//fmt.Println("Call struct or pointer")
		callStructValueMethod(ctx, msg, fvalue)
	default:
		// TODO Reply with MajorError/Fatal ... propagate to parent
	}
}

func callStructMethod(ctx CloveContext, msg Message, s interface{}) {
	callStructValueMethod(ctx, msg, reflect.ValueOf(s))
}

func callStructValueMethod(ctx CloveContext, msg Message, svalue reflect.Value) {
	stype := svalue.Type()
	if stype.Kind() == reflect.Struct {
		stype = reflect.PtrTo(stype)
		ptr := reflect.New(stype)
		ptr.Elem().Set(svalue)
		svalue = ptr
	}
	if stype.Kind() != reflect.Ptr {
		// TODO reply with MajorError/Fatal ... propagte to parent
		return
	}
	ctype := reflect.TypeOf(msg.Content())
	meths := findMethodsOf(svalue, reflect.TypeOf(ctx), ctype)
	if len(meths) == 0 {
		meths = findMethodsOf(svalue, ctype)
	}
	if len(meths) == 0 {
		meths = findMethodsOf(svalue, reflect.TypeOf(ctx), reflect.TypeOf(msg))
	}
	if len(meths) == 0 {
		meths = findMethodsOf(svalue, reflect.TypeOf(msg))
	}
	if len(meths) == 0 {
		// TODO reploy with no method found ... propagate to parent
		return
	}
	callBehaviorValue(ctx, msg, meths[0])
}

func callStructMethodByName(ctx CloveContext, s interface{}, methodName string, params ...interface{}) []interface{} {
	return callStructValueMethodByName(ctx, reflect.ValueOf(s), methodName, params...)
}

func callStructValueMethodByName(ctx CloveContext, svalue reflect.Value, methodName string, params ...interface{}) []interface{} {
	valuesToInterface := func(vals []reflect.Value) []interface{} {
		result := make([]interface{}, len(vals))
		for i, v := range vals {
			result[i] = v.Interface()
		}
		return result
	}

	ptypes := []reflect.Type{reflect.TypeOf(ctx)}
	pvalues := []reflect.Value{reflect.ValueOf(ctx)}
	paramTypes := make([]reflect.Type, len(params))
	paramValues := make([]reflect.Value, len(params))
	for i, param := range params {
		paramTypes[i] = reflect.TypeOf(param)
		paramValues[i] = reflect.ValueOf(param)
	}

	if vmeth, ok := findMethodByName(svalue, methodName, append(ptypes, paramTypes...)...); ok {
		return valuesToInterface(vmeth.Call(append(pvalues, paramValues...)))
	}

	if vmeth, ok := findMethodByName(svalue, methodName, paramTypes...); ok {
		return valuesToInterface(vmeth.Call(paramValues))
	}

	if vmeth, ok := findMethodByName(svalue, methodName, ptypes...); ok {
		return valuesToInterface(vmeth.Call(pvalues))
	}

	return []interface{}{}
}

func defaultLifecycleHandler(runnable CloveRunnable) {
	if err := runnable.PreStart(); err != nil {
		// TODO propagate error to parent with self-ref
		return
	}

	go func() {
		for {
			msg, ok := <-runnable.Messages()
			if !ok {
				// channel closed exit loop
				break
			}
			//fmt.Printf("Got message of %T at %v\n", msg.Content(), runnable.Path())
			ctx := runnable.NewContext(msg.Context())
			switch data := msg.Content(); data.(type) {
			case ChildStoppedEvent:
				evt := data.(ChildStoppedEvent)
				runnable.RemoveChild(evt.Ref)
				msg.Reply(Done())
			case StopCommand:
				//go func() {
				if err := runnable.PreStop(); err != nil {
					// TODO propagate error
				}

				cl := make([]CloveRef, len(runnable.Children()))
				copy(cl, runnable.Children())
				for _, child := range cl {
					<-child.Request(context.Background(), Stop())
					runnable.RemoveChild(child)
				}

				if err := runnable.PostStop(); err != nil {
					// TODO propagte error
				}

				if parent, ok := runnable.Parent(); ok {
					parent.Send(ChildStopped(runnable.Self()))
				}

				msg.Reply(Stopped())

				return // TODO return ends only go func not loop
				//}()
			default:
				if runnable.Clove().Sync {
					// fmt.Println("Execute in sync mode")
					CallBehavior(ctx, msg, runnable.Behavior())
				} else {
					// fmt.Println("Execute in async mode")
					go CallBehavior(ctx, msg, runnable.Behavior())
				}
			}
		}
	}()

	if err := runnable.PostStart(); err != nil {
		fmt.Println("Error while PostStart", runnable.Clove().Name, err)
		// TODO kill self and propagate error to parent
	}
}
