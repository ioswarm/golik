package golik

import "time"

type Timer struct {
	done chan bool
	timer *time.Timer
}

func (t *Timer) Stop() bool {
	defer func() { t.done <- true }()
	return t.timer.Stop()
}

func (t *Timer) Reset(duration time.Duration) bool {
	return t.timer.Reset(duration)
}


type Ticker struct {
	done chan bool
	ticker *time.Ticker
}

func (t *Ticker) Stop() {
	t.ticker.Stop()
	t.done <- true
}

func (t *Ticker) Reset(duration time.Duration) {
	t.ticker.Reset(duration)
}