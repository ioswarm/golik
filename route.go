package golik

import (
	"context"
	"io"
)

type RouteContext interface {
	Loggable
	context.Context
	Handler() CloveHandler

	Header() Values
	Params() Values
	Queries() Values
	Method() string
	Content() io.Reader
}

type Route struct {
	Path      string
	Method    string
	Handle    interface{}
	Subroutes []Route
}

type Response struct {
	Header     Values
	StatusCode int
	Content    interface{}
}

type RouteHandler interface {
	Handle(Route) error
}
