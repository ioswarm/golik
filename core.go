package golik

import (
	"fmt"
	"time"
)


func core(system *golik) *clove {
	c := &clove{
		system: system, 
		name: "core",
		children: make([]CloveRef, 0),
		messages: make(chan Message, 1000), // TODO buffer-size from settings
		receiver: func(context CloveContext) CloveReceiver{
			return &MinionReceiver{
				context: context,
				minion: &coreMinion{},
			}
		},
		runnable: defaultRunnable,
	}

	c.log = newLogrusLogger(map[string]interface{}{
		"name": c.Name(),
		"path": c.Path(),
	})

	c.run()

	return c
}

type coreMinion struct{
	httpRef CloveRef
	pubsub CloveRef
}

func (cm *coreMinion) PostStart(c CloveRef) {
	cm.httpRef = c.Of(Http(), "http")
	cm.pubsub = c.Of(pubsub(), "pubsub")
	
	c.Info("core-system started")
}

func (*coreMinion) PostStop(c CloveRef) {
	c.Info("PostStop core")
}

func (cm *coreMinion) AddRouteProxy(cmd AddRouteCommand, c CloveRef) (Event,error) {
	c.Debug("route %T to %v", cmd, cm.httpRef)
	switch evt := <- cm.httpRef.Ask(cmd, time.Second); evt.(type) { // TODO configure internal timeout
	case error:
		return nil, evt.(error)
	case RouteAddedEvent:
		return evt.(RouteAddedEvent), nil
	default:
		return nil, fmt.Errorf("Received unknown type %T from http-clove", evt)
	}
}