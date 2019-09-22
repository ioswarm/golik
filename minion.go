package golik

import (
	"reflect"
)

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
			runnable: defaultRunnable,
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


// TODO check if minion is pointer !!!
type MinionReceiver struct {
	context GolikContext
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