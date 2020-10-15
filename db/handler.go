package db

import (
	"github.com/ioswarm/golik"
	"github.com/ioswarm/golik/filter"
)

type Handler interface {
	Filter(golik.CloveContext, *filter.Filter) (*filter.Result, error)

	Create(golik.CloveContext, *CreateCommand) error
	Read(golik.CloveContext, *GetCommand) (interface{}, error)
	Update(golik.CloveContext, *UpdateCommand) error
	// TODO Patch
	Delete(golik.CloveContext, *DeleteCommand) (interface{}, error)

	OrElse(ctx golik.CloveContext, msg golik.Message)
}

