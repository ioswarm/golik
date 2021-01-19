package golik

import (
	"context"
	"io"
	"net/http"

	"github.com/gorilla/mux"
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
	Path       string
	Method     string
	Handle     interface{}
	Subroutes  []Route
	Middleware []mux.MiddlewareFunc
}

type ResponseWriter func(HttpRouteContext, http.ResponseWriter, *Response)

type Response struct {
	Header     Values
	StatusCode int
	Content    interface{}
	Writer     ResponseWriter
}

type RouteHandler interface {
	Handle(Route) error
}
