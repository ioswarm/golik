package main

import (
	"time"
	"github.com/ioswarm/golik"
)

type Ticker struct {
	ticker *time.Ticker
}

func (t* Ticker)PostStart(c golik.CloveRef) {
	t.ticker = c.Ticker(2*time.Second, "Ping")
}

func (t* Ticker)PreStop() {
	t.ticker.Stop()
}

func (t* Ticker)PingReceive(s string, c golik.CloveRef) {
	c.Info("Publish %v", s)
	c.Publish(s)
}

type Receiver struct {}

func (r *Receiver)PostStart(c golik.CloveRef) {
	c.Subscribe(c, func(val interface{}) bool {
		switch val.(type) {
		case string: return true
		default: return false
		}
	})
}

func (r* Receiver)Receive(s string, c golik.CloveRef) {
	c.Info("Receive %v", s)
}

// TODO explain in more detail
func main() {
	system := golik.GolikSystem()

	system.Of(golik.CloveDefinition{
		Name: "ticker",
		Receiver: golik.Minion(&Ticker{}),
	})
	
	system.Of(golik.CloveDefinition{
		Name: "receiver",
		Receiver: golik.Minion(&Receiver{}),
	})

	<- system.Terminated()
}