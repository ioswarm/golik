package persistance

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/ioswarm/golik"
)

type HandlerCreation func(golik.CloveContext) (Handler, error)

type ConnectionPoolSettings struct {
	Name       string
	PoolSize   int
	Type       reflect.Type
	IndexField string
	Options    map[string]interface{}

	PreStart  interface{}
	PostStart interface{}
	PreStop   interface{}
	PostStop  interface{}

	CreateHandler HandlerCreation
	Behavior      interface{}
}

func NewConnectionPool(settings *ConnectionPoolSettings) *golik.Clove {
	return &golik.Clove{
		Name:     settings.Name,
		Sync:     true,
		PreStart: settings.PreStart,
		PreStop: settings.PreStop,
		PostStop: settings.PostStop,
		Behavior: func(ctx golik.CloveContext, msg golik.Message) {
			children := make([]golik.CloveRef, len(ctx.Children()))
			copy(children, ctx.Children())
			sort.Slice(children, func(i, j int) bool {
				return children[i].Length() < children[j].Length()
			})

			if len(children) > 0 {
				children[0].Forward(msg)
			}
		},
		PostStart: func(ctx golik.CloveContext) error {
			for i := 0; i < settings.PoolSize; i++ {
				handler, err := settings.CreateHandler(ctx)
				if err != nil {
					return err
				}
				if _, err := ctx.Execute(&golik.Clove{
					Name:     fmt.Sprintf("%v-%v", settings.Name, i),
					Behavior: handler,
				}); err != nil {
					return err
				}
			}

			if settings.PostStart != nil {
				if err := golik.CallLifecycle(ctx, settings.PostStart); err != nil {
					return err
				}
			}

			return nil
		},
	}
}
