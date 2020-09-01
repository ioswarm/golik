package golik

import (
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

type Golik interface {
	Loggable
	CloveExecuter
	Name() string
	At(path string) (*CloveRef, bool)
	Terminate()
	Terminated() <- chan int

	ExecuteService(srv Service) error

	NewTimer(duration time.Duration, f func(time time.Time)) *time.Timer
	NewTicker(interval time.Duration, f func(time time.Time)) *time.Ticker
}

func NewSystem(name string) (Golik, error) {
	initSettings()
	initLogging()

	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	sys := &coreSystem{
		name: name,
		log: logrus.WithFields(logrus.Fields{
			"golik": name,
			"hostname": hostname,
		}),
		exitChan: make(chan int),
	}

	cc, err := newCore().execute(nil, sys)
	if err != nil {
		return nil, err
	}
	sys.core = cc

	usr, err := newUsr().execute(cc, sys)
	if err != nil {
		return nil, err
	}
	sys.usr = usr
	cc.appendChild(usr)

	srv, err := newSrv().execute(cc, sys)
	if err != nil {
		return nil, err
	}
	sys.srv = srv
	cc.appendChild(srv)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<- sigs 
		sys.Terminate()
	}()

	return sys, nil
}

type coreSystem struct {
	name string
	log *logrus.Entry
	exitChan chan int
	core *cloveRunnable
	srv *cloveRunnable
	usr *cloveRunnable
	mutex sync.Mutex
}

func (sys *coreSystem) Name() string {
	return sys.name;
}

func (sys *coreSystem) At(path string) (*CloveRef, bool) {
	if runnable, exists := sys.core.at(path); exists {
		return runnable.Self(), exists
	}
	return nil, false
}

func (sys *coreSystem) Terminate() {
	go func() {
		select {
		case res := <- sys.core.Self().Request(Stop{}):
			switch res.(type) {
			case error:
				sys.Error("Error while stoppping cloves: %v", res)
			case Stopped:
				sys.exitChan <- 0
			}
		case <- time.After(30 * time.Second): // TODO configure termination-timeout
			sys.Error("Timeout while stopping cloves")
			sys.exitChan <- 1
		}
	}()
}

func (sys *coreSystem) Terminated() <- chan int {
	return sys.exitChan
}

func (sys *coreSystem) Run(clove *Clove) (*CloveRef, error) {
	return sys.usr.Run(clove)
}

func (sys *coreSystem) ExecuteService(srv Service) error {
	_, err := sys.srv.Run(srv.CreateInstance(sys))
	return err
}

func (sys *coreSystem) Logger() *logrus.Entry {
	return sys.log
}

func (sys *coreSystem) Log(entry LogEntry) {
	HandleLogEntry(sys.log, entry)
}

func (sys *coreSystem) Debug(msg string, values ...interface{}){
	sys.Log(LogEntry{
		Level: DEBUG,
		Message: msg,
		Values: values,
	})
}

func (sys *coreSystem) Info(msg string, values ...interface{}){
	sys.Log(LogEntry{
		Level: INFO,
		Message: msg,
		Values: values,
	})
}

func (sys *coreSystem) Warn(msg string, values ...interface{}) {
	sys.Log(LogEntry{
		Level: WARN,
		Message: msg,
		Values: values,
	})
}

func (sys *coreSystem) Error(msg string, values ...interface{}) {
	sys.Log(LogEntry{
		Level: ERROR,
		Message: msg,
		Values: values,
	})
}

func (sys *coreSystem) Panic(msg string, values ...interface{}) {
	sys.Log(LogEntry{
		Level: PANIC,
		Message: msg,
		Values: values,
	})
}

func (sys *coreSystem) NewTimer(duration time.Duration, f func(time time.Time)) *time.Timer {
	t := time.NewTimer(duration)

	go func() {
		f(<- t.C)
	}()

	return t
}

func (sys *coreSystem) NewTicker(interval time.Duration, f func(time time.Time)) *time.Ticker {
	t := time.NewTicker(interval)

	go func() {
		f(<- t.C)
	}()

	return t
}



func newCore() *Clove {
	return &Clove{
		Name: "core",
		Receive: func(ctx CloveContext) func(msg Message) {
			return func(msg Message) {

			}
		},
	}
}

func newSrv() *Clove {
	return &Clove{
		Name: "srv",
		Receive: func(ctx CloveContext) func(msg Message) {
			return func(msg Message) {

			}
		},
	}
}

func newUsr() *Clove {
	return &Clove{
		Name: "usr",
		Receive: func(ctx CloveContext) func(msg Message) {
			return func(msg Message) {

			}
		},
	}
}