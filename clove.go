package golik

import (
	"time"
	"strings"
)

type Producer func(CloveRef, string) CloveRef

type CloveReceiver interface {
	Receive(reference CloveRef, messages <-chan Message)
}


type CloveRoute func (CloveRef) *Route

type CloveContext interface {
	GolikContext

	System() Golik

	Children() []CloveRef
	Child(name string) (CloveRef, bool)
}

type CloveRef interface {
	CloveContext
	Name() string
	Path() string

	Root() CloveRef
	Parent() CloveRef
	
	Channel() chan<- Message
	Tell(interface{})
	Ask(interface{}, time.Duration) <-chan interface{}

	ChannelSize() int 

	WithRoute(CloveRoute) error

	Timer(duration time.Duration, message interface{}) *time.Timer
	Ticker(duration time.Duration, message interface{}) *time.Ticker
}

type clove struct {
	// TODO implement state
	system Golik
	parent   CloveRef
	name     string
	children []CloveRef
	messages chan Message
	receiver func(CloveContext) CloveReceiver
	runnable func(*clove)
	log Loggable
}

/* constructors */

func (c *clove) run() {
	c.runnable(c)
}

/* internal implementation */



/* GolikContext implementation */

func (c *clove) Of(producer Producer, name string) CloveRef {
	// TODO validate name and return error
	if ref, ok := c.Child(name); !ok{
		ref = producer(c, name)
		c.appendChild(ref)
		// TODO check Client state
		return ref
	} else {
		c.Debug("Clove %v already exists", name)
		return ref // TODO maybe return error???
	}
}

func (c *clove) At(path string) (CloveRef, bool) {
	if (len(path) > 0) {
		if path[0] == '/' {
			if len(path) == 1 {
				return c.Root(), true
			}
			return c.Root().At(path[1:])
		} else if path[0] == '.' || strings.HasPrefix(path, c.Name()) {
			if i := strings.IndexRune(path, '/'); i > -1 {
				if i < len(path)-1 {
					return c.At(path[i+1:])
				}
			} else {
				return c, true
			}
		} else if len(path) >= 2  && path[0:2] == ".." && c.parent != nil {
			if len(path) == 2 || path == "../" {
				return c.parent, true
			}
			return c.parent.At(path[3:])
		} else {
			if i := strings.IndexRune(path, '/'); i != -1 {
				child, ok := c.Child(path[:i])
				if !ok {
					return nil, false
				}
				if i < len(path)-1 {
					return child.At(path[i+1:])
				}
				return child, true
			}
			return c.Child(path)
		}
	}

	return nil, false
}

func (c *clove) System() Golik {
	return c.system
}

func (c *clove) Publish(msg interface{}) {
	c.System().Publish(msg)
}

func (c *clove) Subscribe(ref CloveRef, f func(interface{}) bool) {
	c.System().Subscribe(ref, f)
}

func (c *clove) Unsubscribe(ref CloveRef) {
	c.System().Unsubscribe(ref)
}

func (c *clove) Children() []CloveRef {
	return c.children
}

func (c *clove) Child(name string) (CloveRef, bool) {
	for _, child := range c.children {
		if name == child.Name() {
			return child, true
		}
	}
	return nil, false
}

func (c *clove) containsChild(ref CloveRef) bool {
	return c.indexOfChild(ref) >= 0
}

func (c *clove) appendChild(ref CloveRef) bool {
	if contains := c.containsChild(ref); !contains {
		c.Debug("Append clove %v to children", ref.Name())
		c.children = append(c.children, ref)
		return true
	}
	return false
}

func (c *clove) indexOfChild(ref CloveRef) int  {
	for i, child := range c.children {
		if child.Path() == ref.Path() {
			return i
		}
	}
	return -1
}

func (c *clove) removeChild(ref CloveRef) bool {
	return c.removeChildAt(c.indexOfChild(ref))
}

func (c *clove) removeChildAt(index int) bool {
	if index >= 0 && index < len(c.children) {
		c.Debug("Remove clove at %v", index)
		c.children = append(c.children[:index], c.children[index+1:]...)
	}
	return false
}

/* CloveRef implementation */

func (c *clove) Name() string {
	return c.name
}

func (c *clove) Path() string {
	pathSeg := "/" + c.Name()
	if c.parent != nil {
		return c.parent.Path() + pathSeg
	}
	return pathSeg
}

func (c *clove) Root() CloveRef {
	if c.parent != nil {
		return c.Parent().Root()
	}
	return c
}

func (c *clove) Parent() CloveRef {
	return c.parent
}

func (c *clove) Channel() chan<- Message {
	return c.messages
}

func (c *clove) Tell(message interface{}) {
	c.messages <- NewMessage(message)
}

func (c *clove) Ask(message interface{}, timeout time.Duration) <-chan interface{} {
	// TODO check state of clove 
	return await(message, c.Channel(), timeout)
}

func (c *clove) ChannelSize() int {
	return len(c.messages)
}


func (c *clove) WithRoute(f CloveRoute) error {
	if err := c.System().WithRoute(f(c)); err != nil {
		return err
	}
	return nil
}

func (c *clove) Timer(duration time.Duration, message interface{}) *time.Timer {
	timer := time.NewTimer(duration)
	
	go func() {
		<- timer.C
		c.Tell(message)
	}()

	return timer
}

func (c *clove) Ticker(duration time.Duration, message interface{}) *time.Ticker {
	ticker := time.NewTicker(duration)

	go func() {
		for range ticker.C {
			c.Tell(message)
		}
	}()

	return ticker
}

/* Loggable implementation */

func (c *clove) Debug(msg string, values ...interface{}){
	c.log.Debug(msg, values...)
}
func (c *clove) Info(msg string, values ...interface{}){
	c.log.Info(msg, values...)
}
func (c *clove) Warn(msg string, values ...interface{}){
	c.log.Warn(msg, values...)
}
func (c *clove) Error(msg string, values ...interface{}){
	c.log.Error(msg, values...)
}
func (c *clove) Panic(msg string, values ...interface{}){
	c.log.Panic(msg, values...)
}