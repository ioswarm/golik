package golik

import (
	"fmt"
	"sync"
)

type Subscribe struct {
	Ref    CloveRef
	Filter func(interface{}) bool
}

type Unsubscribe struct {
	Ref CloveRef
}

type Publish struct {
	Content interface{}
}

type core struct {
	subscriptions []Subscribe
	mutex         sync.Mutex
}

func newCore() *Clove {
	return &Clove{
		Name: "core",
		Behavior: &core{
			subscriptions: make([]Subscribe, 0),
		},
		PreStart: func(ctx CloveContext) {
			fmt.Println("\033[1;36m", title, "\033[0m")
		},
	}
}

func (c *core) addSubsciption(s Subscribe) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.subscriptions = append(c.subscriptions, s)
}

func (c *core) removeSubscription(ref CloveRef) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	for i, s := range c.subscriptions {
		if s.Ref.Path() == ref.Path() {
			c.subscriptions = append(c.subscriptions[:i], c.subscriptions[i+1:]...)
		}
	}
}

func (c *core) Messages(ctx CloveContext, msg Message) {
	if msg.Content() == nil {
		return
	}
	switch content := msg.Content(); content.(type) {
	case Subscribe:
		c.addSubsciption(content.(Subscribe))
		msg.Reply(Done())
	case Unsubscribe:
		us := content.(Unsubscribe)
		c.removeSubscription(us.Ref)
		msg.Reply(Done())
	case Publish:
		pub := content.(Publish)
		for _, s := range c.subscriptions {
			if s.Filter(pub.Content) {
				s.Ref.Send(pub.Content)
			}
		}
		msg.Reply(Done())
	}
}
