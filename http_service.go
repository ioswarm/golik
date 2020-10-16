package golik

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
)

type HttpService struct {
	name     string
	system   Golik
	server   *http.Server
	settings *httpSettings
	handler CloveHandler
	Router   *mux.Router
}

func Http(system Golik) (*HttpService, error) {
	return NewHttp("http", system)
}

func NewHttp(name string, system Golik) (*HttpService, error) {
	hs := &HttpService{
		name:   name,
		system: system,
		settings: newHTTPSettings(name),
		Router: mux.NewRouter(),
	}

	con, err := system.ExecuteService(hs)
	if err != nil {
		return nil, err
	}

	hs.handler = con

	return hs, nil
}

func (hs *HttpService) CreateServiceInstance(system Golik) *Clove {
	return &Clove{
		Name: hs.name,
		Behavior: func(ctx CloveContext, msg Message) {
			msg.Reply(Done())  // 
		},
		PreStart: func(ctx CloveContext) {
			hs.run(ctx)
		},
		PostStop: func(ctx CloveContext) {
			hs.shutdown(ctx)
		},
	}
}

func (hs *HttpService) run(ctx CloveContext) {
	hs.server = &http.Server{
		Addr:         hs.settings.Addr(),
		Handler:      hs.Router,
		ReadTimeout:  hs.settings.ReadTimeout,
		WriteTimeout: hs.settings.WriteTimeout,
		IdleTimeout:  hs.settings.IdleTimeout,
	}

	go func() {
		if err := hs.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			ctx.Error("Error in http-service execution ...", err)
		}
	}()

	ctx.Info("Http-Server '%v' is listening on %v", hs.name, hs.settings.Addr())
}

func (hs *HttpService) shutdown(ctx CloveContext) error {
	c, cancel := context.WithTimeout(context.Background(), hs.settings.ShutdownTimeout)
	defer cancel()

	hs.server.SetKeepAlivesEnabled(false)
	if err := hs.server.Shutdown(c); err != nil {
		ctx.Warn("Could not gracefully shutdown Http-Server '%v': %v\n", hs.name, err)
		return err
	}
	ctx.Info("Http-Server '%v' at '%v' shutdown ...", hs.name, hs.settings.Addr())
	return nil
}

func (hs *HttpService) HandleFunc(path string, f func(http.ResponseWriter, *http.Request)) *mux.Route {
	return hs.Router.HandleFunc(path, f)
}

func (hs *HttpService) Handle(route Route) error {
	return hs.handleRoute(hs.Router, route)
}

func (hs *HttpService) handleRoute(mrouter *mux.Router, route Route) error {
	if route.Handle == nil && len(route.Subroutes) == 0 {
		return Errorf("Handle-Func and Subroutes are empty for %v", route.Path)
	}

	method := "GET"
	if route.Method != "" {
		method = route.Method
	}

	//r := mrouter.NewRoute().Path(route.Path)
	r := mrouter.NewRoute().Path(route.Path)
	if len(route.Subroutes) > 0 {
		r = mrouter.PathPrefix(route.Path)
	}
	if route.Handle != nil {
		if err := validateRouteFunc(route.Handle); err != nil {
			return err
		}
		
		r.Methods(method).HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := newHttpRouteContext(r.Context(), hs.handler, r)
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