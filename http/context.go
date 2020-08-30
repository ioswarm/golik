package http

import (
	"io"
	ht "net/http"

	"github.com/ioswarm/golik"
	"github.com/sirupsen/logrus"
	"github.com/gorilla/mux"
)

type httpRouteContext struct {
	system golik.Golik
	log *logrus.Entry
	request *ht.Request
}

func (ctx *httpRouteContext) System() golik.Golik {
	return ctx.system
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

func (ctx *httpRouteContext) Logger() *logrus.Entry {
	return ctx.log
}

func (ctx *httpRouteContext) Log(entry golik.LogEntry) {
	golik.HandleLogEntry(ctx.log, entry)
}

func (ctx *httpRouteContext) Debug(msg string, values ...interface{}){
	ctx.Log(golik.LogEntry{
		Level: golik.DEBUG,
		Message: msg,
		Values: values,
	})
}

func (ctx *httpRouteContext) Info(msg string, values ...interface{}){
	ctx.Log(golik.LogEntry{
		Level: golik.INFO,
		Message: msg,
		Values: values,
	})
}

func (ctx *httpRouteContext) Warn(msg string, values ...interface{}) {
	ctx.Log(golik.LogEntry{
		Level: golik.WARN,
		Message: msg,
		Values: values,
	})
}

func (ctx *httpRouteContext) Error(msg string, values ...interface{}) {
	ctx.Log(golik.LogEntry{
		Level: golik.ERROR,
		Message: msg,
		Values: values,
	})
}

func (ctx *httpRouteContext) Panic(msg string, values ...interface{}) {
	ctx.Log(golik.LogEntry{
		Level: golik.PANIC,
		Message: msg,
		Values: values,
	})
}