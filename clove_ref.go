package golik

import (
	"context"
)

type CloveRef interface {
	Name() string
	Path() string
	Capacity() int
	Length() int

	Send(interface{})
	Request(context.Context, interface{}) <-chan interface{}
	RequestFunc(context.Context, interface{}) (interface{}, error)
	Forward(Message)
}

type ref struct {
	name string
	path string 
	messages chan<- Message
}

func newRef(runnable *cloveRunnable) CloveRef {
	return &ref{
		name: runnable.clove.Name,
		path: runnable.Path(),
		messages: runnable.messages,
	}
}

func (r *ref) Name() string {
	return r.name
}

func (r *ref) Path() string {
	return r.path
}

func (r *ref) Capacity() int {
	return cap(r.messages)
}

func (r *ref) Length() int {
	return len(r.messages)
}

func (r *ref) Send(data interface{}) {
	r.messages <- newMessage(context.Background(), data)
}

func (r *ref) Request(ctx context.Context, data interface{}) <-chan interface{} {
	m := newMessage(ctx, data)
	r.messages <- m
	return m.Result()
}

func (r *ref) RequestFunc(ctx context.Context, data interface{}) (interface{}, error) {
	result := r.Request(ctx, data)
	select {
	case <-ctx.Done():
		// DO not wait <-result
		return nil, ctx.Err()
	case res := <-result:
		switch res.(type) {
		case error:
			return nil, res.(error)
		default:
			return res, nil
		}
	}
}

func (r *ref) Forward(msg Message) {
	r.messages <- msg
}