package golik

import (
	"fmt"
	"reflect"
	"sort"
)

type HandlerCreation func(CloveContext) (Handler, error)

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

type ConnectionPool interface {
	CreateConnectionPool(*ConnectionPoolSettings) (CloveRef, error)
}

func NewConnectionPool(settings *ConnectionPoolSettings) *Clove {
	return &Clove{
		Name:     settings.Name,
		Sync:     true,
		PreStart: settings.PreStart,
		PreStop:  settings.PreStop,
		PostStop: settings.PostStop,
		Behavior: func(ctx CloveContext, msg Message) {
			children := make([]CloveRef, len(ctx.Children()))
			copy(children, ctx.Children())
			sort.Slice(children, func(i, j int) bool {
				return children[i].Length() < children[j].Length()
			})

			if len(children) > 0 {
				children[0].Forward(msg)
			}
		},
		PostStart: func(ctx CloveContext) error {
			size := settings.PoolSize
			if size == 0 {
				size = 10 // TODO configure default poolsize
			}

			for i := 0; i < size; i++ {
				handler, err := settings.CreateHandler(ctx)
				if err != nil {
					return err
				}
				if _, err := ctx.Execute(&Clove{
					Name:     fmt.Sprintf("%v-%v", settings.Name, i),
					Behavior: handler,
				}); err != nil {
					return err
				}
			}

			if settings.PostStart != nil {
				if err := CallLifecycle(ctx, settings.PostStart); err != nil {
					return err
				}
			}

			return nil
		},
	}
}
