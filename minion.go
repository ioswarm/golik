package golik

import (
	"reflect"
	"sort"
	ior "github.com/ioswarm/goutils/reflect"
)

/*
func Minion(f func() interface{}) Producer {

	return func(parent CloveRef, name string) CloveRef {
		obj := f()
		c := &clove{
			system: parent.System(),
			parent: parent,
			name: name,
			children: make([]CloveRef, 0),
			messages: make(chan Message, 1000), // TODO buffer-size from settings
			receiver: func(context CloveContext) CloveReceiver{
				return &MinionReceiver{
					context: context,
					minion: obj,
				}
			},
			runnable: DefaultRunnable,
		}

		c.log = newLogrusLogger(map[string]interface{}{
			"name": c.Name(),
			"path": c.Path(),
			"minion": reflect.TypeOf(obj).String(),
		})

		c.run()

		return c
	}
}
*/

/*
func Minion(f func() interface{}) func(CloveContext) CloveReceiver {
	return func(context CloveContext) CloveReceiver{
		return &MinionReceiver{
			minion: f(),
		}
	}
}*/

func Minion(obj interface{}) func(CloveContext) CloveReceiver {
	return func(context CloveContext) CloveReceiver{
		return &MinionReceiver{
			minion: obj,
		}
	}
}

// TODO check if minion is pointer !!!
type MinionReceiver struct {
	minion interface{}
}

func (r *MinionReceiver) Receive(reference CloveRef, messages <-chan Message) {
	go func() {
		defer reference.Debug("Receiver messaging loop ended")
		for {
			msg, ok := <- messages
			if !ok {
				reference.Debug("Receiver channel is closed, no more messages will be processed")
				return
			}
			
			if result, err := callMinionLogic(reference, r.minion, msg.Payload()); err != nil {
				msg.Reply(err)
			} else if result != nil {
				msg.Reply(result)
			} else {
				reference.Debug("No receiver-result for '%T' reply 'Nothing'", msg.Payload())
				msg.Reply(Nothing())
			}
			
		}
	}()
}

/* lifecycle proxies */

func (r *MinionReceiver) PreStart(c CloveRef) {
	callLifeCycle(c, r.minion, "PreStart")
}

func (r *MinionReceiver) PostStart(c CloveRef) {
	callLifeCycle(c, r.minion, "PostStart")

	// TODO if one or more routes bind ... minion clove is not stoppable or find a way to remove route from mux.Router
	// or rebuild mux.Router and restart http-server
	c.Debug("Bind route-methods to golik")
	for _, route := range callMinionRoutes(r.minion) {
		if err := c.System().WithRoute(route); err != nil {
			c.Warn("Could not bind route %v: %v", route, err.Error())
		}
	}
	// TODO call minion CloveRoutes
	c.Debug("Bind clove-routes to golik")
	for _, cloveRoute := range callMinionCloveRoutes(r.minion) {
		if err := c.WithRoute(cloveRoute); err != nil {
			c.Warn("Could not bind clove-route: %v", err.Error())
		}
	}
}

func (r *MinionReceiver) PreStop(c CloveRef) {
	callLifeCycle(c, r.minion, "PreStop")
}

func (r *MinionReceiver) PostStop(c CloveRef) {
	callLifeCycle(c, r.minion, "PostStop")
}



// reflect
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