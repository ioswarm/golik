package golik

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

type CloveExecuter interface {
	Run(clove *Clove) (*CloveRef, error)
}

type CloveContext interface {
	Loggable
	CloveExecuter
	System() Golik
	Parent() (*CloveRef, bool)
	Self() *CloveRef
	Children() []*CloveRef
	Child(name string) (*CloveRef, bool)

	Stop()
}

type CloveRunnableContext interface {
	CloveContext
	Clove() *Clove
	Messages() <-chan Message

	RemoveChild(child *CloveRef) bool
	SetTimer(timer *time.Timer)
	StopTimer()
}

type ReceiveFunc func(ctx CloveContext) func(msg Message)
type LifecycleFunc func(ctx CloveContext)

type Clove struct {
	Name       string
	Receive    ReceiveFunc
	Handler    HandlerFunc
	BufferSize uint32
	Async      bool
	Timeout    time.Duration
	RefrestTimeout bool

	PreStart  LifecycleFunc
	PostStart LifecycleFunc
	PreStop   LifecycleFunc
	PostStop  LifecycleFunc
}

func Folder(name string, children ...*Clove) *Clove {
	return EmptyClove(name, children...)
}

func EmptyClove(name string, children ...*Clove) *Clove {
	return &Clove{
		Name: name,
		Receive: func(ctx CloveContext) func(msg Message) {

			for _, child := range children {
				// TODO log errors
				ctx.Run(child)
			}

			return func(msg Message) {
				// DO NOTHING
			}
		},
	}
}

func EmptyReceive() ReceiveFunc {
	return func(ctx CloveContext) func(msg Message) {
		return func(msg Message) {
			// DO NOTHING
		}
	}
}

func (c *Clove) execute(parent *cloveRunnable, system Golik) (*cloveRunnable, error) {
	if c.Name == "" {
		return nil, errors.New("Clove's name is not defined") // TODO create and use name-generator for empty names; like docker
	}
	if c.Receive == nil {
		return nil, errors.New("Receiver-Function is not defined")
	}
	if c.Handler == nil {
		c.Handler = defaultHandler
	}
	if c.BufferSize == 0 {
		c.BufferSize = 1000 // TODO configure default buffer-size (in golik.clove) and buffer-size per clove eg. golik.clove.{path-segments}
	}

	runnable := &cloveRunnable{
		system:   system,
		parent:   parent,
		clove:    c,
		messages: make(chan Message, c.BufferSize),
	}
	runnable.log = system.Logger().WithFields(logrus.Fields{
		"clove": c.Name,
		"path":  runnable.path(),
	})

	c.Handler(runnable)

	return runnable, nil
}

type cloveRunnable struct {
	system       Golik
	parent       *cloveRunnable
	clove        *Clove
	children     []*cloveRunnable
	messages     chan Message
	log          *logrus.Entry
	mutex        sync.Mutex
	timeoutTimer *time.Timer
}

func (c *cloveRunnable) at(path string) (*cloveRunnable, bool) {
	if len(path) > 0 {
		if path[0] == '/' {
			if len(path) == 1 {
				return c.root(), true
			}
			return c.root().at(path[1:])
		} else if path[0] == '.' || strings.HasPrefix(path, c.clove.Name) {
			if i := strings.IndexRune(path, '/'); i > -1 {
				if i < len(path)-1 {
					return c.at(path[i+1:])
				}
			} else {
				return c, true
			}
		} else if len(path) >= 2 && path[0:2] == ".." && c.parent != nil {
			if len(path) == 2 || path == "../" {
				return c.parent, true
			}
			return c.parent.at(path[3:])
		} else {
			if i := strings.IndexRune(path, '/'); i != -1 {
				child, ok := c.child(path[:i])
				if !ok {
					return nil, false
				}
				if i < len(path)-1 {
					return child.at(path[i+1:])
				}
				return child, true
			}
			return c.child(path)
		}
	}

	return nil, false
}

func (c *cloveRunnable) path() string {
	if c.parent == nil {
		return "/"
	}
	ppath := c.parent.path()
	if strings.HasSuffix(ppath, "/") {
		return ppath + c.clove.Name
	}
	return ppath + "/" + c.clove.Name
}

func (c *cloveRunnable) Clove() *Clove {
	return c.clove
}

func (c *cloveRunnable) Messages() <-chan Message {
	return c.messages
}

func (c *cloveRunnable) System() Golik {
	return c.system
}

func (c *cloveRunnable) root() *cloveRunnable {
	if c.parent != nil {
		return c.parent.root()
	}
	return c
}

func (c *cloveRunnable) Parent() (*CloveRef, bool) {
	if c.parent != nil {
		return c.parent.Self(), true
	}
	return nil, false
}

func (c *cloveRunnable) Self() *CloveRef {
	return &CloveRef{
		name:     c.clove.Name,
		messages: c.messages,
		path:     c.path(),
		executer: c,
	}
}

func (c *cloveRunnable) Children() []*CloveRef {
	result := make([]*CloveRef, len(c.children))
	for i, child := range c.children {
		result[i] = child.Self()
	}
	return result
}

func (c *cloveRunnable) child(name string) (*cloveRunnable, bool) {
	for _, child := range c.children {
		if child.clove.Name == name {
			return child, true
		}
	}
	return nil, false
}

func (c *cloveRunnable) Child(name string) (*CloveRef, bool) {
	if child, exists := c.child(name); exists {
		return child.Self(), exists
	}
	return nil, false
}

func (c *cloveRunnable) indexOfChild(path string) int {
	for i, child := range c.children {
		if child.path() == path {
			return i
		}
	}
	return -1
}

func (c *cloveRunnable) containsChild(cr *cloveRunnable) bool {
	return c.indexOfChild(cr.path()) >= 0
}

func (c *cloveRunnable) appendChild(cr *cloveRunnable) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if contains := c.containsChild(cr); !contains {
		c.Debug("Append clove '%v' to children", cr.clove.Name)
		c.children = append(c.children, cr)
		return true
	}
	return false
}

func (c *cloveRunnable) RemoveChild(ref *CloveRef) bool {
	return c.removeChildAt(c.indexOfChild(ref.Path()))
}

func (c *cloveRunnable) removeChildAt(index int) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if index >= 0 && index < len(c.children) {
		c.children = append(c.children[:index], c.children[index+1:]...)
	}
	return false
}

func (c *cloveRunnable) Stop() {
	c.Self().Tell(Stop{})
}

func (c *cloveRunnable) SetTimer(timer *time.Timer) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.timeoutTimer != nil {
		c.timeoutTimer.Stop()
	}

	c.timeoutTimer = timer
}

func (c *cloveRunnable) StopTimer() {
	if c.timeoutTimer != nil {
		c.mutex.Lock()
		defer c.mutex.Unlock()

		c.timeoutTimer.Stop()
	}
}

func (c *cloveRunnable) Run(clove *Clove) (*CloveRef, error) {
	if clove == nil {
		return nil, errors.New("Clove is nil")
	}

	for _, child := range c.children {
		if clove.Name == child.clove.Name {
			return nil, fmt.Errorf("Clove '%v' already exists", clove.Name)
		}
	}

	cc, err := clove.execute(c, c.system)
	if err != nil {
		return nil, err
	}
	c.appendChild(cc)

	return cc.Self(), nil
}

func (c *cloveRunnable) Logger() *logrus.Entry {
	return c.log
}

func (c *cloveRunnable) Log(entry LogEntry) {
	HandleLogEntry(c.log, entry)
}

func (c *cloveRunnable) Debug(msg string, values ...interface{}) {
	c.Log(LogEntry{
		Level:   DEBUG,
		Message: msg,
		Values:  values,
	})
}

func (c *cloveRunnable) Info(msg string, values ...interface{}) {
	c.Log(LogEntry{
		Level:   INFO,
		Message: msg,
		Values:  values,
	})
}

func (c *cloveRunnable) Warn(msg string, values ...interface{}) {
	c.Log(LogEntry{
		Level:   WARN,
		Message: msg,
		Values:  values,
	})
}

func (c *cloveRunnable) Error(msg string, values ...interface{}) {
	c.Log(LogEntry{
		Level:   ERROR,
		Message: msg,
		Values:  values,
	})
}

func (c *cloveRunnable) Panic(msg string, values ...interface{}) {
	c.Log(LogEntry{
		Level:   PANIC,
		Message: msg,
		Values:  values,
	})
}

type CloveRef struct {
	name     string
	path     string
	messages chan Message
	executer CloveExecuter
}

func (cr *CloveRef) Name() string {
	return cr.name
}

func (cr *CloveRef) Path() string {
	return cr.path
}

func (cr *CloveRef) Length() int {
	return len(cr.messages)
}

func (cr *CloveRef) Capacity() int {
	return cap(cr.messages)
}

func (cr *CloveRef) Tell(payload interface{}) {
	m := NewMessage(cr, payload)
	cr.messages <- m
}

func (cr *CloveRef) Ask(payload interface{}, timeout time.Duration) <-chan interface{} {
	result := make(chan interface{}, 1)

	go func() {
		m := NewMessage(cr, payload)
		cr.messages <- m
		select {
		case res := <-m.Result():
			result <- res
			close(result)
		case <-time.After(timeout):
			result <- errors.New("Timeout") // TODO
		}
	}()

	return result
}

func (cr *CloveRef) AskFunc(payload interface{}, timeout time.Duration) (interface{}, error) {
	switch result := <- cr.Ask(payload, timeout); result.(type) {
	case error:
		return nil, result.(error)
	default:
		return result, nil
	}
}

func (cr *CloveRef) Request(payload interface{}) <-chan interface{} {
	m := NewMessage(cr, payload)
	cr.messages <- m
	return m.Result()
}

func (cr *CloveRef) RequestFunc(payload interface{}) (interface{}, error) {
	switch result := <- cr.Request(payload); payload.(type) {
	case error: 
		return nil, result.(error)
	default:
		return result, nil
	}
}

func (cr *CloveRef) Forward(msg Message) {
	cr.messages <- msg
}

func (cr *CloveRef) Run(clove *Clove) (*CloveRef, error) {
	return cr.executer.Run(clove)
}
