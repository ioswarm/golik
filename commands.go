package golik


// Lifecycle commands

type StopCommand struct {}

func Stop() StopCommand {
	return StopCommand{}
}


type DoneEvent struct{}

func Done() DoneEvent {
	return DoneEvent{}
}

type StoppedEvent struct {}

func Stopped() StoppedEvent {
	return StoppedEvent{}
}

type ChildStoppedEvent struct {
	Ref CloveRef
}

func ChildStopped(ref CloveRef) ChildStoppedEvent {
	return ChildStoppedEvent{
		Ref: ref,
	}
}




// data pool commands

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