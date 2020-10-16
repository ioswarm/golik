package golik

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type CloveExecutor interface {
	At(string) (CloveRef, bool)
	Execute(clove *Clove) (CloveRef, error)
}

type Golik interface {
	CloveExecutor
	Name() string
	Terminate()
	TerminateWithTimeout(time.Duration)
	Terminated() <-chan int

	Settings() Settings

	ExecuteService(Service) (CloveHandler, error)

	NewTimer(time.Duration, func(time.Time)) *Timer
	NewTicker(time.Duration, func(time.Time)) *Ticker
}

func NewSystem(name string) (Golik, error) {
	sys := &golikSystem{
		name:     name,
		exitChan: make(chan int, 1),
		settings: NewSettings(),
	}

	sys.core = newRunnable(sys, nil, newCore())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigs
		// TODO log this
		sys.Terminate()
	}()

	sys.core.run()

	usr, err := sys.core.executeInternal(newUsr())
	if err != nil {
		return nil, err
	}
	sys.usr = usr

	srv, err := sys.core.executeInternal(newSrv())
	if err != nil {
		return nil, err
	}
	sys.srv = srv

	return sys, nil
}

type golikSystem struct {
	name     string
	exitChan chan int
	settings Settings
	core     *cloveRunnable
	srv      *cloveRunnable
	usr      *cloveRunnable
}

func (sys *golikSystem) Name() string {
	return sys.name
}

func (sys *golikSystem) Terminate() {
	sys.TerminateWithTimeout(sys.settings.TerminationTimeout())
}

func (sys *golikSystem) TerminateWithTimeout(timeout time.Duration) {
	go func() {
		select {
		case res := <-sys.core.Self().Request(context.Background(), Stop()):
			switch res.(type) {
			case error:
				log.Fatalf("Error while stoppping cloves: %v", res)
			case StoppedEvent:
				log.Println("golik is going down bye")
				sys.exitChan <- 0
			}
		case <-time.After(timeout):
			log.Fatalln("Timeout while stopping cloves")
		}
	}()
}

func (sys *golikSystem) Terminated() <-chan int {
	return sys.exitChan
}

func (sys *golikSystem) Settings() Settings {
	return sys.settings
}

func (sys *golikSystem) Execute(clove *Clove) (CloveRef, error) {
	return sys.usr.Execute(clove)
}

func (sys *golikSystem) At(path string) (CloveRef, bool) {
	return sys.core.At(path)
}

func (sys *golikSystem) ExecuteService(srv Service) (CloveHandler, error) {
	runnable, err := sys.srv.executeInternal(srv.CreateServiceInstance(sys))
	if err != nil {
		return nil, err
	}
	return runnable, nil
}

func (sys *golikSystem) NewTimer(duration time.Duration, f func(time time.Time)) *Timer {
	t := time.NewTimer(duration)
	done := make(chan bool)

	go func() {
		select {
		case <-done:
			return
		case tx := <-t.C:
			f(tx)
		}
	}()

	return &Timer{
		done:  done,
		timer: t,
	}
}

func (sys *golikSystem) NewTicker(interval time.Duration, f func(time time.Time)) *Ticker {
	t := time.NewTicker(interval)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case tx := <-t.C:
				f(tx)
			}
		}
	}()

	return &Ticker{
		done:   done,
		ticker: t,
	}
}
