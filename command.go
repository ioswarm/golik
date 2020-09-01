package golik

type Done struct {  }

type Stop struct {  }
type Stopped struct{ }

type StopChild struct { }

type ChildStopped struct {
	Child *CloveRef
}

type Timeout struct {}
