package http

import (
	"context"
	"encoding/json"
	ht "net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/ioswarm/golik"
	"github.com/sirupsen/logrus"
)

type HttpService struct {
	name     string
	system   golik.Golik
	log      *logrus.Entry
	server   *ht.Server
	settings *httpSettings
	Router   *mux.Router
}

func Http() *HttpService {
	return NewHttp("http")
}

func NewHttp(name string) *HttpService {
	return &HttpService{
		name:   name,
		Router: mux.NewRouter(),
	}
}

func (hs *HttpService) CreateInstance(system golik.Golik) *golik.Clove {
	return &golik.Clove{
		Name: hs.name,
		Receive: func(ctx golik.CloveContext) func(msg golik.Message) {

			return func(msg golik.Message) {

			}
		},
		PreStart: func(ctx golik.CloveContext) {
			hs.run(ctx)
		},
		PreStop: func(ctx golik.CloveContext) {
			hs.shutdown(ctx)
		},
	}
}

func (hs *HttpService) run(ctx golik.CloveContext) {
	hs.system = ctx.System()
	hs.log = ctx.Logger()
	settings := newHTTPSettings(hs.name)
	hs.settings = settings
	hs.server = &ht.Server{
		Addr:         settings.Addr(),
		Handler:      hs.Router,
		ReadTimeout:  settings.ReadTimeout,
		WriteTimeout: settings.WriteTimeout,
		IdleTimeout:  settings.IdleTimeout,
	}

	go func() {
		if err := hs.server.ListenAndServe(); err != nil && err != ht.ErrServerClosed {
			ctx.Error("Error in http-service execution ...", err)
		}
	}()

	ctx.Info("Http-Server '%v' is listening on %v", hs.name, settings.Addr())
}

func (hs *HttpService) shutdown(ctx golik.CloveContext) error {
	c, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	hs.server.SetKeepAlivesEnabled(false)
	if err := hs.server.Shutdown(c); err != nil {
		ctx.Warn("Could not gracefully shutdown Http-Server '%v': %v\n", hs.name, err)
		return err
	}
	ctx.Info("Http-Server '%v' at '%v' shutdown ...", hs.name, hs.settings.Addr())
	return nil
}

func (hs *HttpService) Handle(route golik.Route) error {
	if err := validateRouteFunc(route.Handle); err != nil {
		return err
	}

	method := "GET"
	if route.Method != "" {
		method = route.Method
	}

	hs.Router.HandleFunc(route.Path, func(w ht.ResponseWriter, r *ht.Request) {
		ctx := &httpRouteContext{
			system: hs.system,
			log: hs.log.WithFields(logrus.Fields{
				"httpPath":   route.Path,
				"httpMethod": route.Method,
			}),
			request: r,
		}

		resp := handleRoute(ctx, route.Handle)

		for key := range resp.Header {
			w.Header().Add(key, resp.Header.Get(key))
		}
		w.Header().Add("Content-Type", "application/json; utf-8")
		w.WriteHeader(resp.StatusCode)

		if resp.Content != nil {
			enc := json.NewEncoder(w)
			switch resp.Content.(type) {
			case golik.Error, *golik.Error:
				if err := enc.Encode(resp.Content); err != nil {
					ctx.Warn(err.Error())
				}
			case error:
				if err := enc.Encode(golik.NewError(resp.Content.(error))); err != nil {
					ctx.Warn(err.Error())
				}
			default:
				if err := enc.Encode(resp.Content); err != nil {
					ctx.Warn(err.Error())
				}
			}
		}
	}).Methods(method)

	return nil
}


