package golik

import (
	"errors"
	"fmt"
	"time"
)

type GolikContext interface {
	Loggable
	// Of(producer Producer, name string) CloveRef
	Of(definition CloveDefinition) CloveRef
	At(path string) (CloveRef, bool)

	Publish(msg interface{})
	Subscribe(c CloveRef, f func(interface{}) bool)
	Unsubscribe(c CloveRef)
}

type Golik interface {
	GolikContext
	Terminate() <-chan bool
	Terminated() <-chan bool

	WithRoute(route *Route) error
}

func GolikSystem() Golik {
	InitViperSettings()

	system := &golik{
		// TODO settings
		quitChannel: make(chan bool, 1),
		log: newLogrusLogger(map[string]interface{}{}),
	}

	system.core = core(system)

	return system
}

type golik struct {
	core *clove
	quitChannel chan bool
	log Loggable
}

/*
func (g *golik) Of(producer Producer, name string) CloveRef {
	return g.core.Of(producer, name)
}
*/

func (g *golik) Of(definition CloveDefinition) CloveRef {
	return g.core.Of(definition)
}

func (g *golik) At(path string) (CloveRef, bool) {
	return g.core.At(path)
}

func (g *golik) Terminate() <-chan bool {
	go func(){
		defer close(g.quitChannel)
		switch res := <- g.core.Ask(Stop(), 10 * time.Second); res.(type) { // TODO configure stop-timeout
		case error: 
			g.Panic("Error while termination: %v", res)
		default:
			g.Info("Shutdown golik ... Bye")
			g.quitChannel <- true	
		}
	}()
	return g.quitChannel
}

func (g *golik) Terminated() <-chan bool {
	return g.quitChannel
}

func (g *golik) WithRoute(route *Route) error {
	if route == nil {
		return errors.New("Given Route is nil")
	}
	switch evt := <- g.core.Ask(AddRoute(route), time.Second); evt.(type) { // TODO configure internal timeout
	case error:
		return evt.(error)
	case RouteAddedEvent:
		return nil
	default:
		return fmt.Errorf("Receive unknown result %T, while adding route %v", evt, route)
	}
}

func (g *golik) pubsub() (CloveRef, bool) {
	return g.At("/core/pubsub")
}

func (g *golik) Publish(msg interface{}) {
	if ref, ok := g.pubsub(); ok {
		ref.Tell(msg)
	}
}

func (g *golik) Subscribe(c CloveRef, f func(interface{}) bool) {
	if ref, ok := g.pubsub(); ok {
		ref.Tell(Subscribe(c, f))
	}
}

func (g *golik) Unsubscribe(c CloveRef) {
	if ref, ok := g.pubsub(); ok {
		ref.Tell(Unsubscribe(c))
	}
}

func (g *golik) Debug(msg string, values ...interface{}){
	g.log.Debug(msg, values...)
}
func (g *golik) Info(msg string, values ...interface{}){
	g.log.Info(msg, values...)
}
func (g *golik) Warn(msg string, values ...interface{}){
	g.log.Warn(msg, values...)
}
func (g *golik) Error(msg string, values ...interface{}){
	g.log.Error(msg, values...)
}
func (g *golik) Panic(msg string, values ...interface{}){
	g.log.Panic(msg, values...)
}