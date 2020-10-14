package golik

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

type LifecycleHandler func(CloveRunnable)

type CloveHandler interface {
	Loggable
	CloveExecutor
	System() Golik
	Self() CloveRef
	Parent() (CloveRef, bool)
	Children() []CloveRef
	Child(string) (CloveRef, bool)

	Path() string

	Publish(interface{})
	Subscribe(func(interface{}) bool) error
	Unsubscribe()
}

type CloveRunnable interface {
	CloveHandler
	Clove() *Clove
	Settings() CloveSettings

	Messages() <-chan Message

	Behavior() interface{}

	NewContext(context.Context) CloveContext
	NewContextWithTimeout(context.Context, time.Duration) CloveContext
	NewContextWithDeadline(context.Context, time.Time) CloveContext

	PreStart() error
	PostStart() error
	PreStop() error
	PostStop() error

	RemoveChild(CloveRef) bool
}

func newRunnable(system Golik, parent *cloveRunnable, clove *Clove) *cloveRunnable {
	handler := clove.LifecycleHandler
	if clove.LifecycleHandler == nil {
		handler = defaultLifecycleHandler
	}

	name := clove.Name
	if name == "" {
		name = hash()
	}

	settings := system.Settings().CloveSettings(name)

	runnable := &cloveRunnable{
		system:          system,
		parent:          parent,
		children:        make([]*cloveRunnable, 0),
		clove:           clove,
		name:            name,
		settings:        settings,
		messages:        make(chan Message, settings.BufferSize()),
		currentBehavior: clove.Behavior,
		handler:         handler,
	}

	return runnable
}

type cloveRunnable struct {
	system          Golik
	parent          *cloveRunnable
	children        []*cloveRunnable
	clove           *Clove
	name            string
	settings        CloveSettings
	messages        chan Message
	currentBehavior interface{}
	handler         LifecycleHandler
	mutex           sync.Mutex
}

func (cr *cloveRunnable) run() {
	cr.handler(cr)
}

func (cr *cloveRunnable) System() Golik {
	return cr.system
}

func (cr *cloveRunnable) Self() CloveRef {
	return newRef(cr)
}

func (cr *cloveRunnable) root() *cloveRunnable {
	if cr.parent != nil {
		return cr.parent.root()
	}
	return cr
}

func (cr *cloveRunnable) Parent() (CloveRef, bool) {
	if cr.parent != nil {
		return cr.parent.Self(), true
	}
	return nil, false
}

func (cr *cloveRunnable) Children() []CloveRef {
	result := make([]CloveRef, 0)
	for _, child := range cr.children {
		result = append(result, child.Self())
	}
	return result
}

func (cr *cloveRunnable) child(name string) (*cloveRunnable, bool) {
	for _, child := range cr.children {
		if child.Clove().Name == name {
			return child, true
		}
	}
	return nil, false
}

func (cr *cloveRunnable) Child(name string) (CloveRef, bool) {
	if child, ok := cr.child(name); ok {
		return child.Self(), ok
	}
	return nil, false
}

func (cr *cloveRunnable) appendChild(child *cloveRunnable) bool {
	cr.mutex.Lock()
	defer cr.mutex.Unlock()

	if _, ok := cr.Child(child.Clove().Name); !ok {
		cr.children = append(cr.children, child)
		return true
	}
	return false
}

func (cr *cloveRunnable) RemoveChild(child CloveRef) bool {
	cr.mutex.Lock()
	defer cr.mutex.Unlock()

	for i, chd := range cr.children {
		if chd.Path() == child.Path() {
			cr.children = append(cr.children[:i], cr.children[i+1:]...)
			return true
		}
	}

	return false
}

func (cr *cloveRunnable) Clove() *Clove {
	return cr.clove
}

func (cr *cloveRunnable) Settings() CloveSettings {
	return cr.settings
}

func (cr *cloveRunnable) Messages() <-chan Message {
	return cr.messages
}

func (cr *cloveRunnable) Behavior() interface{} {
	return cr.currentBehavior
}

func (cr *cloveRunnable) executeInternal(clove *Clove) (*cloveRunnable, error) {
	if clove == nil {
		return nil, Errorln("Clove is not defined")
	}
	if err := clove.Validate(); err != nil {
		return nil, err
	}

	child := newRunnable(cr.System(), cr, clove)
	if cr.appendChild(child) {
		child.run()
		return child, nil
	}

	return nil, fmt.Errorf("Child clove with name '%v' alraedy exists", clove.Name)
}

func (cr *cloveRunnable) Execute(clove *Clove) (CloveRef, error) {
	r, err := cr.executeInternal(clove)
	if err != nil {
		return nil, err
	}
	return r.Self(), nil
}

func (cr *cloveRunnable) NewContext(parent context.Context) CloveContext {
	context, cancel := context.WithCancel(parent)
	return newCloveContext(context, cancel, cr)
}

func (cr *cloveRunnable) NewContextWithTimeout(parent context.Context, timeout time.Duration) CloveContext {
	context, cancel := context.WithTimeout(parent, timeout)
	return newCloveContext(context, cancel, cr)
}

func (cr *cloveRunnable) NewContextWithDeadline(parent context.Context, d time.Time) CloveContext {
	context, cancel := context.WithDeadline(parent, d)
	return newCloveContext(context, cancel, cr)
}

func (cr *cloveRunnable) Path() string {
	if cr.parent == nil {
		return "/"
	}
	ppath := cr.parent.Path()
	if strings.HasSuffix(ppath, "/") {
		return ppath + cr.clove.Name
	}
	return ppath + "/" + cr.clove.Name
}

func (cr *cloveRunnable) at(path string) (*cloveRunnable, bool) {
	if len(path) > 0 {
		if path[0] == '/' {
			if len(path) == 1 {
				return cr.root(), true
			}
			return cr.root().at(path[1:])
		} else if path[0] == '.' || strings.HasPrefix(path, cr.clove.Name) {
			if i := strings.IndexRune(path, '/'); i > -1 {
				if i < len(path)-1 {
					return cr.at(path[i+1:])
				}
			} else {
				return cr, true
			}
		} else if len(path) >= 2 && path[0:2] == ".." && cr.parent != nil {
			if len(path) == 2 || path == "../" {
				return cr.parent, true
			}
			return cr.parent.at(path[3:])
		} else {
			if i := strings.IndexRune(path, '/'); i != -1 {
				child, ok := cr.child(path[:i])
				if !ok {
					return nil, false
				}
				if i < len(path)-1 {
					return child.at(path[i+1:])
				}
				return child, true
			}
			return cr.child(path)
		}
	}

	return nil, false
}

func (cr *cloveRunnable) At(path string) (CloveRef, bool) {
	if runnable, ok := cr.at(path); ok {
		return runnable.Self(), ok
	}
	return nil, false
}

func (cr *cloveRunnable) Publish(data interface{}) {
	cr.root().Self().Send(Publish{Content: data})
}

func (cr *cloveRunnable) Subscribe(f func(interface{}) bool) error {
	subscribe := Subscribe{
		Ref:    cr.Self(),
		Filter: f,
	}
	ctx := cr.NewContextWithTimeout(context.Background(), cr.settings.SubscriptionTimeout())
	if _, err := cr.root().Self().RequestFunc(ctx, subscribe); err != nil {
		return err
	}
	return nil
}

func (cr *cloveRunnable) Unsubscribe() {
	cr.root().Self().Send(Unsubscribe{
		Ref: cr.Self(),
	})
}

func (cr *cloveRunnable) PreStart() error {
	ctx := cr.NewContextWithTimeout(context.Background(), cr.Settings().PreStartTimeout())
	if cr.Clove().PreStart != nil {
		if err := CallLifecycle(ctx, cr.Clove().PreStart); err != nil {
			return err
		}
	}
	result := callStructMethodByName(ctx, cr.Behavior(), "PreStart")
	if len(result) > 0 {
		res := result[0]
		switch res.(type) {
		case error:
			return res.(error)
		}
	}
	return nil
}

func (cr *cloveRunnable) PostStart() error {
	ctx := cr.NewContextWithTimeout(context.Background(), cr.Settings().PostStartTimeout())
	if cr.Clove().PostStart != nil {
		if err := CallLifecycle(ctx, cr.Clove().PostStart); err != nil {
			return err
		}
	}
	result := callStructMethodByName(ctx, cr.Behavior(), "PostStart")
	if len(result) > 0 {
		res := result[0]
		switch res.(type) {
		case error:
			return res.(error)
		}
	}
	return nil
}

func (cr *cloveRunnable) PreStop() error {
	ctx := cr.NewContextWithTimeout(context.Background(), cr.Settings().PreStopTimeout())
	if cr.Clove().PreStop != nil {
		if err := CallLifecycle(ctx, cr.Clove().PreStop); err != nil {
			return err
		}
	}
	result := callStructMethodByName(ctx, cr.Behavior(), "PreStop")
	if len(result) > 0 {
		res := result[0]
		switch res.(type) {
		case error:
			return res.(error)
		}
	}
	return nil
}

func (cr *cloveRunnable) PostStop() error {
	ctx := cr.NewContextWithTimeout(context.Background(), cr.Settings().PostStopTimeout())
	if cr.Clove().PostStop != nil {
		if err := CallLifecycle(ctx, cr.Clove().PostStop); err != nil {
			return err
		}
	}
	result := callStructMethodByName(ctx, cr.Behavior(), "PostStop")
	if len(result) > 0 {
		res := result[0]
		switch res.(type) {
		case error:
			return res.(error)
		}
	}
	return nil
}

func (cr *cloveRunnable) Log(log *LogEntry) {
	log.WithMeta(map[string]string{
		"system": cr.System().Name(),
		"path":   cr.Path(),
	})
	cr.Publish(log)
}

func (cr *cloveRunnable) Debug(msg string, values ...interface{}) {
	cr.Log(NewLogDebug(msg, values...))
}

func (cr *cloveRunnable) Info(msg string, values ...interface{}) {
	cr.Log(NewLogInfo(msg, values...))
}

func (cr *cloveRunnable) Warn(msg string, values ...interface{}) {
	cr.Log(NewLogWarn(msg, values...))
}

func (cr *cloveRunnable) Error(msg string, values ...interface{}) {
	cr.Log(NewLogError(msg, values...))
}
