package golik

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