package http

import (
	"context"
	"encoding/json"
	"fmt"
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

func Http(system golik.Golik) (*HttpService, error) {
	return NewHttp("http", system)
}

func NewHttp(name string, system golik.Golik) (*HttpService, error) {
	hs := &HttpService{
		name:   name,
		system: system,
		Router: mux.NewRouter(),
	}

	if err := system.ExecuteService(hs); err != nil {
		return nil, err
	}

	return hs, nil
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

func (hs *HttpService) System() golik.Golik {
	return hs.system
}

func (hs *HttpService) run(ctx golik.CloveContext) {
	//hs.system = ctx.System()
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

func (hs *HttpService) HandleFunc(path string, f func(ht.ResponseWriter, *ht.Request)) *mux.Route {
	return hs.Router.HandleFunc(path, f)
}

func (hs *HttpService) Handle(route golik.Route) error {
	return hs.handleRoute(hs.Router, route)
}


func (hs *HttpService) handleRoute(mrouter *mux.Router, route golik.Route) error {
	if route.Handle == nil && len(route.Subroutes) == 0 {
		return fmt.Errorf("Handle-Func and Subroutes are empty for %v", route.Path)
	}

	method := "GET"
	if route.Method != "" {
		method = route.Method
	}

	//r := mrouter.NewRoute().Path(route.Path)
	r := mrouter.PathPrefix(route.Path)
	if route.Handle != nil {
		if err := validateRouteFunc(route.Handle); err != nil {
			return err
		}
		
		r.Methods(method).HandlerFunc(func(w ht.ResponseWriter, r *ht.Request) {
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
				if err := json.NewEncoder(w).Encode(resp.Content); err != nil {
					ctx.Warn(err.Error())
				}
			}
		})
	}

	if len(route.Subroutes) > 0 {
		srouter := r.Subrouter()
		for _, sr := range route.Subroutes {
			if err := hs.handleRoute(srouter, sr); err != nil {
				return err
			}
		}
	}

	return nil
}
