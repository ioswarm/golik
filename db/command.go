package db

type CreateCommand struct {
	Entity interface{} `json:"entity,omitempty"`
}

func Create(entity interface{}) *CreateCommand {
	return &CreateCommand{Entity: entity}
}

type GetCommand struct {
	Id interface{} `json:"id,omitempty"`
}

func Get(id interface{}) *GetCommand {
	return &GetCommand{Id: id}
}

type UpdateCommand struct {
	Id     interface{} `json:"id,omitempty"`
	Entity interface{} `json:"entity,omitempty"`
}

func Update(id interface{}, entity interface{}) *UpdateCommand {
	return &UpdateCommand{Id: id, Entity: entity}
}

type DeleteCommand struct {
	Id interface{} `json:"id,omitempty"`
}

func Delete(id interface{}) *DeleteCommand {
	return &DeleteCommand{Id: id}
}
