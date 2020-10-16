package golik

type Handler interface {
	Filter(CloveContext, *Filter) (*Result, error)

	Create(CloveContext, *CreateCommand) error
	Read(CloveContext, *GetCommand) (interface{}, error)
	Update(CloveContext, *UpdateCommand) error
	// TODO Patch
	Delete(CloveContext, *DeleteCommand) (interface{}, error)

	OrElse(ctx CloveContext, msg Message)
}

