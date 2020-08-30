package golik

import (
	"io"
)

type Values map[string]string

func (v Values) Add(key string, value string) {
	v[key] = value
}

func (v Values) Get(key string) string {
	if v == nil {
		return ""
	}
	return v[key]
}

type RouteContext interface {
	Loggable
	System() Golik
	Header() Values
	Params() Values
	Queries() Values
	Method() string
	Content() io.Reader
}

type Route struct {
	Path string
	Method string
	Handle interface{}
}

type Response struct {
	Header Values
	StatusCode int
	Content interface{}
}

type RouteHandler interface {
	Handle(Route) error
}
