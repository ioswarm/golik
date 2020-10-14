package golik

import (
	"context"
	"sync"
)

type CloveContext interface {
	CloveHandler
	context.Context
	Cancel()
}

type cloveContext struct {
	context.Context
	runnable CloveRunnable
	cancel   context.CancelFunc
	options map[string]interface{}

	mutex sync.Mutex
}

func newCloveContext(context context.Context, cancel context.CancelFunc, runnable CloveRunnable) CloveContext {
	return &cloveContext{context, runnable, cancel, make(map[string]interface{}), sync.Mutex{}}
}

func (cc *cloveContext) System() Golik {
	return cc.runnable.System()
}

func (cc *cloveContext) Self() CloveRef {
	return cc.runnable.Self()
}

func (cc *cloveContext) Parent() (CloveRef, bool) {
	return cc.runnable.Parent()
}

func (cc *cloveContext) Children() []CloveRef {
	return cc.runnable.Children()
}

func (cc *cloveContext) Child(name string) (CloveRef, bool) {
	return cc.runnable.Child(name)
}

func (cc *cloveContext) Cancel() {
	cc.cancel()
}

func (cc *cloveContext) Execute(clove *Clove) (CloveRef, error) {
	return cc.runnable.Execute(clove)
}

func (cc *cloveContext) Path() string {
	return cc.runnable.Path()
}

func (cc *cloveContext) At(path string) (CloveRef, bool) {
	return cc.runnable.At(path)
}

func (cc *cloveContext) Publish(data interface{}) {
	cc.runnable.Publish(data)
}

func (cc *cloveContext) Subscribe(f func(interface{}) bool) error {
	return cc.runnable.Subscribe(f)
}

func (cc *cloveContext) Unsubscribe() {
	cc.runnable.Unsubscribe()
}

func (cc *cloveContext) AddOption(name string, value interface{}) {
	cc.mutex.Lock()
	defer cc.mutex.Unlock()
	cc.options[name] = value
}

func (cc *cloveContext) Option(name string) (interface{}, bool) {
	if value, ok := cc.options[name]; ok {
		return value, ok
	}
	return nil, false
}

func (cc *cloveContext) Debug(msg string, values ...interface{}) {
	cc.runnable.Debug(msg, values...)
}

func (cc *cloveContext) Info(msg string, values ...interface{}) {
	cc.runnable.Info(msg, values...)
}

func (cc *cloveContext) Warn(msg string, values ...interface{}) {
	cc.runnable.Warn(msg, values...)
}

func (cc *cloveContext) Error(msg string, values ...interface{}) {
	cc.runnable.Error(msg, values...)
}