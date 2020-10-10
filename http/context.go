package http

import (
	"context"
	"io"
	ht "net/http"

	"github.com/gorilla/mux"
	"github.com/ioswarm/golik"
)

func newHttpRouteContext(context context.Context, handler golik.CloveHandler, request *ht.Request) *httpRouteContext {
	return &httpRouteContext{context, handler, request}
}

type httpRouteContext struct {
	context.Context
	handler golik.CloveHandler
	request *ht.Request
}

func (ctx *httpRouteContext) Handler() golik.CloveHandler {
	return ctx.handler
}

func (ctx *httpRouteContext) Header() golik.Values {
	header := make(map[string]string)
	uheader := ctx.request.Header
	for k := range uheader {
		header[k] = uheader.Get(k)
 	}
	return header
}

func (ctx *httpRouteContext) Params() golik.Values {
	return mux.Vars(ctx.request)
}

func (ctx *httpRouteContext) Queries() golik.Values {
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

