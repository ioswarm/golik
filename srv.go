package golik

type Service interface {
	CreateServiceInstance(system Golik) *Clove
}

func newSrv() *Clove {
	return &Clove{
		Name: "srv",
		Behavior: func(msg Message) {
			msg.Reply(Errorln("Sending to srv is not supported"))
		},
		PostStart: func(ctx CloveContext) error {
			if _, err := ctx.Execute(NewLogger()); err != nil {
				return err
			}

			ctx.Info("%v is up ... ready to receive new services", ctx.Path())
			return nil
		},
	}
}