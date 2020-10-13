package persistance

import (
	"fmt"
	"sort"

	"github.com/ioswarm/golik"
)

type ConnectionPoolSettings interface {
	Name() string

	PoolSize() int

	Connect(golik.CloveContext) error
	Close(golik.CloveContext) error

	CreateHandler(golik.CloveContext) (Handler, error)
}

func NewConnectionPool(settings ConnectionPoolSettings) *golik.Clove {
	return &golik.Clove{
		Name:     settings.Name(),
		Sync:     true,
		PreStart: settings.Connect,
		PostStop: settings.Close,
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
			for i := 0; i < settings.PoolSize(); i++ {
				handler, err := settings.CreateHandler(ctx)
				if err != nil {
					return err
				}
				if _, err := ctx.Execute(&golik.Clove{
					Name:     fmt.Sprintf("%v-%v", settings.Name(), i),
					Behavior: handler,
				}); err != nil {
					return err
				}
			}
			return nil
		},
	}
}
