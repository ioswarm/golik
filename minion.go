package golik

import (
	"reflect"
	"sync"
	//"sort"
)

type MinionHandler interface {
	CallLifeCycle(string, CloveContext)
	HandleReceive(ctx CloveContext) func(Message)
}

type MinionConfig struct {
	Name string
	Async bool
	BufferSize uint32
	Stateful bool
	Handler MinionHandler
}

func Stateful(obj interface{}, config MinionConfig) *Clove {
	return Minion(obj, MinionConfig{
		Name: config.Name,
		BufferSize: config.BufferSize,
		Async: config.Async,
		Stateful: true,
		Handler: config.Handler,
	})
}

func Stateless(obj interface{}, config MinionConfig) *Clove {
	return Minion(obj, MinionConfig{
		Name: config.Name,
		BufferSize: config.BufferSize,
		Async: true,
		Stateful: false,
		Handler: config.Handler,
	})
}

func Minion(obj interface{}, conf MinionConfig) *Clove {
	name := conf.Name
	if obj != nil && name == "" {
		if otype := reflect.TypeOf(obj); otype.Kind() == reflect.Ptr {
			name = otype.Elem().Name()
		} else {
			name = otype.Name()
		}
	}

	mHandler := conf.Handler
	if mHandler == nil {
		mHandler = newMinionHandler(obj, &conf)	
	}

	return &Clove{
		Name: name,
		Receive: mHandler.HandleReceive,
		PreStart: func(ctx CloveContext) {
			mHandler.CallLifeCycle("PreStart", ctx)
		},
		PostStart: func(ctx CloveContext) {
			mHandler.CallLifeCycle("PostStart", ctx)
		},
		PreStop: func(ctx CloveContext) {
			mHandler.CallLifeCycle("PreStop", ctx)
		},
		PostStop: func(ctx CloveContext) {
			mHandler.CallLifeCycle("PostStop", ctx)
		},
		Async: conf.Async,
		BufferSize: conf.BufferSize,
	}
}

func newMinionHandler(obj interface{}, conf *MinionConfig) *minionHandler {
	return &minionHandler{
		minion: obj,
		conf: conf,
	}
}

type minionHandler struct {
	minion interface{}
	conf *MinionConfig
	mutex sync.Mutex
}

func (mh *minionHandler) CallLifeCycle(methodName string, ctx CloveContext) {
	CallLifeCycle(mh.minion, methodName, ctx)
}

func (mh *minionHandler) HandleReceive(ctx CloveContext) func(Message) {
	//ctxValue := reflect.ValueOf(ctx)
	return func(msg Message) {
		if mh.minion == nil {
			return
		}
		if mh.conf.Stateful {
			mh.mutex.Lock()
			defer mh.mutex.Unlock()
		}

		if result, ok := CallMethod(mh.minion, ctx, msg.Payload); ok {
			msg.Reply(result)
		}

		/*pvalue := reflect.ValueOf(msg.Payload)

		meths := utils.FindMethodsOf(mh.minionValue, pvalue.Type())
		meths = append(meths, utils.FindMethodsOf(mh.minionValue, pvalue.Type(), ctxValue.Type())...)
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

			// TODO ... debug info
			meth := meths[0]
			switch methodCategory(meth) {
			case 0:
				ctx.Debug("Method-state-0")
				result := meth.Call([]reflect.Value{pvalue, ctxValue})
				if err, ok := result[1].Interface().(error); ok {
					msg.Reply(err)
				}
				msg.Reply(result[0].Interface())
			case 1:
				ctx.Debug("Method-state-1")
				result := meth.Call([]reflect.Value{pvalue})
				if err, ok := result[1].Interface().(error); ok {
					msg.Reply(err)
				}
				msg.Reply(result[0].Interface())
			case 2:
				ctx.Debug("Method-state-2")
				result := meth.Call([]reflect.Value{pvalue, ctxValue})
				if err, ok := result[0].Interface().(error); ok {
					msg.Reply(err)
				}
				msg.Reply(result[1].Interface())
			case 3:
				ctx.Debug("Method-state-3")
				result := meth.Call([]reflect.Value{pvalue})
				if err, ok := result[0].Interface().(error); ok {
					msg.Reply(err)
				}
				msg.Reply(result[1].Interface())
			case 4:
				ctx.Debug("Method-state-4")
				result := meth.Call([]reflect.Value{pvalue, ctxValue})
				if err, ok := result[0].Interface().(error); ok {
					msg.Reply(err)
				}
			case 5:
				ctx.Debug("Method-state-5")
				result := meth.Call([]reflect.Value{pvalue})
				if err, ok := result[0].Interface().(error); ok {
					msg.Reply(err)
				}
			case 6:
				ctx.Debug("Method-state-6")
				result := meth.Call([]reflect.Value{pvalue, ctxValue})
				msg.Reply(result[0].Interface())
			case 7:
				ctx.Debug("Method-state-7")
				result := meth.Call([]reflect.Value{pvalue})
				msg.Reply(result[0].Interface())
			case 8:
				ctx.Debug("Method-state-8")
				meth.Call([]reflect.Value{pvalue, ctxValue})
			case 9:
				ctx.Debug("Method-state-9")
				meth.Call([]reflect.Value{pvalue})
			default:
				ctx.Debug("Method-state-? not match")
				// TODO throw warning ... method found but too many result-types
			}
		}*/
	}
}