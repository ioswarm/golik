package golik

import (
	"time"
	"context"
	"net/http"
	"github.com/gorilla/mux"
)

/*
func Http() Producer {
	return func(parent CloveRef, name string) CloveRef {
		settings := newBaseSettings()
		if name != "http" {
			settings = newSettings(name)
		}

		c := &clove{
			system: parent.System(),
			parent: parent,
			name: name,
			children: make([]CloveRef, 0),
			messages: make(chan Message), // TODO buffer-size from settings
			receiver: func(context CloveContext) CloveReceiver{
				return &MinionReceiver{
					context: context,
					minion: &HttpService{
						name: name,
						settings: settings,
						router: mux.NewRouter(),
						//routes: make([]*Route, 0),
					},
				}
			},
			runnable: DefaultRunnable,
		}

		c.log = newLogrusLogger(map[string]interface{}{
			"name": c.Name(),
			"path": c.Path(),
			"host": settings.Host,
			"port": settings.Port,
		})

		c.run()

		return c
	}
}
*/

func Http(name string) CloveDefinition {
	settings := newBaseSettings()
	if name != "http" {
		settings = newSettings(name)
	}

	return CloveDefinition{
		Name: name,
		LogParams: map[string]interface{}{
			"host": settings.Host,
			"port": settings.Port,
		},
		Receiver: func(context CloveContext) CloveReceiver {
			return &MinionReceiver{
				minion: &HttpService{
					name: name,
					settings: settings,
					router: mux.NewRouter(),
					//routes: make([]*Route, 0),
				},
			}
		},
	}
}

type HttpService struct {
	name string
	ref CloveRef
	settings *httpSettings
	server *http.Server
	router *mux.Router
	//routes []*Route
}

func (hs *HttpService) PreStart(ref CloveRef) {
	hs.ref = ref
	hs.run()
}

func (hs *HttpService) PostStop() {
	hs.stop()
} 

func (hs *HttpService) run() {
	hs.server = &http.Server{
		Addr: hs.settings.Addr(),
		Handler: hs.router,
		ReadTimeout: hs.settings.ReadTimeout,
		WriteTimeout: hs.settings.WriteTimeout,
		IdleTimeout: hs.settings.IdleTimeout,
	}

	go func(){
		if err := hs.server.ListenAndServe(); err != nil  && err != http.ErrServerClosed {
			hs.ref.Error("Error in http-service execution ...", err)
		}
	}()

	hs.ref.Info("Http-Server '%v' is listening on %v", hs.name, hs.settings.Addr())
}

func (hs *HttpService) stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	hs.server.SetKeepAlivesEnabled(false)
	if err := hs.server.Shutdown(ctx); err != nil {
		hs.ref.Error("Could not gracefully shutdown Http-Server '%v': %v\n", hs.name, err)
	} else {
		hs.ref.Info("Http-Server '%v' shutdown ...", hs.name)
	}
}

func (hs *HttpService) AddRoute(cmd AddRouteCommand) Event { // TODO return error as well
	r := cmd.Route()
	//hs.routes = append(hs.routes, r)
	// add route to mux router
	hs.router.HandleFunc(r.path, func(w http.ResponseWriter, rq *http.Request){
		r.handler(&RouteContext{
			HttpClove: hs.ref,
			Request: rq,
		})(w)
	}).Methods(r.methods...)

	return RouteAdded()
}