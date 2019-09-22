package golik


type Command interface{
	Payload() interface{}
}

/* Commands */

func Stop() Command {
	return StopCommand{}
}

func parentalStop(sender CloveRef) Command {
	return StopCommand{sender: sender}
}

type StopCommand struct {
	sender CloveRef
}

func (cmd StopCommand) Payload() interface{} {
	return cmd.sender
}


func AddRoute(route *Route) Command {
	return AddRouteCommand{
		route: route,
	}
}

type AddRouteCommand struct {
	route *Route
}

func (r AddRouteCommand) Payload() interface{} {
	return r.route
}

func (r AddRouteCommand) Route() *Route {
	return r.route
}


func Subscribe(ref CloveRef, f func(msg interface{}) bool) SubscribeCommand {
	return SubscribeCommand{
		Subscription: Subscription{
			Clove: ref,
			Filter: f,
		},
	}
}

type SubscribeCommand struct {
	Subscription Subscription
}

func (s SubscribeCommand) Payload() interface{} {
	return s.Subscription
}


func Unsubscribe(ref CloveRef) UnsubscribeCommand {
	return UnsubscribeCommand{
		Clove: ref,
	}
}

type UnsubscribeCommand struct {
	Clove CloveRef
}

func (us UnsubscribeCommand) Payload() interface{} {
	return us.Clove
}


type Event interface {
	Command
}

/* Events */

func Stopped() Event {
	return StoppedEvent{}
}

type StoppedEvent struct {}

func (StoppedEvent) Payload() interface{} {
	return nil
}


func ChildStopped(sender CloveRef) Event {
	return ChildStoppedEvent{sender: sender}
}

type ChildStoppedEvent struct{
	sender CloveRef
}

func (evt ChildStoppedEvent) Payload() interface{} {
	return evt.sender
}


func RouteAdded() Event {
	return RouteAddedEvent{}
}

type RouteAddedEvent struct {}

func (RouteAddedEvent) Payload() interface{} {
	return nil
}


func Nothing() Event {
	return NothingEvent{}
}

type NothingEvent struct {}

func (NothingEvent) Payload() interface{} {
	return nil
}