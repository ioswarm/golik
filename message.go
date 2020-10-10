package golik

import "context"

type Message interface {
	Context() context.Context

	Content() interface{}
	Reply(interface{})

	Result() <-chan interface{}
}

func newMessage(ctx context.Context, data interface{}) Message {
	return msg{
		context: ctx,
		content: data,
		reply: make(chan interface{}, 1),
	}
}

type msg struct {
	context context.Context
	content interface{}
	reply chan interface{}
}

func (m msg) Context() context.Context {
	return m.context
}

func (m msg) Content() interface{} {
	return m.content
}

func (m msg) Reply(result interface{}) {
	m.reply <- result
	close(m.reply)
}

func (m msg) Result() <-chan interface{} {
	return m.reply
}