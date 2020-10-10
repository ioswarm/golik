package golik

type Clove struct {
	Name       string
	BufferSize uint32
	Sync       bool

	Behavior interface{}

	LifecycleHandler LifecycleHandler

	PreStart  interface{}
	PostStart interface{}
	PreStop   interface{}
	PostStop  interface{}
}

func (clove *Clove) Validate() error {
	if err := checkBehavior(clove.Behavior); err != nil {
		return err
	}
	if err := checkLifecycleFunc(clove.PreStart); err != nil {
		return Errorln("PreStart is not valid", err)
	}
	if err := checkLifecycleFunc(clove.PostStart); err != nil {
		return Errorln("PostStart is not valid", err)
	}
	if err := checkLifecycleFunc(clove.PreStop); err != nil {
		return Errorln("PreStop is not valid", err)
	}
	if err := checkLifecycleFunc(clove.PostStop); err != nil {
		return Errorln("PostStop is not valid", err)
	}
	return nil
}

