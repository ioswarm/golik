package golik

import (
	"context"
	"io"
	"net/http"

	"github.com/gorilla/mux"
)

type HttpRouteContext interface {
	RouteContext
	Request() *http.Request
}

func newHttpRouteContext(context context.Context, handler CloveHandler, request *http.Request) HttpRouteContext {
	return &httpRouteContext{context, handler, request}
}

type httpRouteContext struct {
	context.Context
	handler CloveHandler
	request *http.Request
}

func (ctx *httpRouteContext) Request() *http.Request {
	return ctx.request
}

func (ctx *httpRouteContext) Handler() CloveHandler {
	return ctx.handler
}

func (ctx *httpRouteContext) Header() Values {
	header := make(map[string]string)
	uheader := ctx.request.Header
	for k := range uheader {
		header[k] = uheader.Get(k)
 	}
	return header
}

func (ctx *httpRouteContext) Params() Values {
	return mux.Vars(ctx.request)
}

func (ctx *httpRouteContext) Queries() Values {
	queries := make(map[string]string)
	uqueriey := ctx.request.URL.Query()
	for k := range uqueriey {
		queries[k] = uqueriey.Get(k)
	}
	return queries
}

func (ctx *httpRouteContext) Method() string {
	mth := ctx.request.Method
	if mth == "" {
		return "GET"
	}
	return mth
}

func (ctx *httpRouteContext) Content() io.Reader {
	return ctx.request.Body
}

func (ctx *httpRouteContext) Debug(msg string, values ...interface{}) {
	ctx.handler.Debug(msg, values...)
}

func (ctx *httpRouteContext) Info(msg string, values ...interface{}) {
	ctx.handler.Info(msg, values...)
}

func (ctx *httpRouteContext) Warn(msg string, values ...interface{}) {
	ctx.handler.Warn(msg, values...)
}

func (ctx *httpRouteContext) Error(msg string, values ...interface{}) {
	ctx.handler.Error(msg, values...)
}

