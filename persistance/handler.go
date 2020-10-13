package persistance

import (
	"github.com/ioswarm/golik"
	"github.com/ioswarm/golik/filter"
)

type Handler interface {
	Filter(golik.CloveContext, filter.Filter) (filter.Result, error)

	Create(golik.CloveContext, Create) error
	Read(golik.CloveContext, Get) (interface{}, error)
	Update(golik.CloveContext, Update) error
	// TODO Patch
	Delete(golik.CloveContext, Delete) (interface{}, error)

	OrElse(golik.CloveContext, msg golik.Message)
}

