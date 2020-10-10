package golik

func newUsr() *Clove {
	return &Clove{
		Name: "usr",
		Behavior: func(msg Message) {
			msg.Reply(Errorln("Sending to usr is not supported"))
		},
		PostStart: func(ctx CloveContext) {
			ctx.Info("%v is up, ready to receive custom cloves", ctx.Path())
		},
		PreStop: func(ctx CloveContext) {
			ctx.Info("%v is going down ... stop all child cloves", ctx.Path())
		},
		PostStop: func(ctx CloveContext) {
			ctx.Info("%v is down", ctx.Path())
		},
	}
}