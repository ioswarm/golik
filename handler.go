package golik

type HandlerFunc func(ctx CloveRunnableContext)

func defaultHandler(ctx CloveRunnableContext) {
	ctx.Debug("PreStart '%v'", ctx.Self().Name())
	if ctx.Clove().PreStart != nil {
		ctx.Clove().PreStart(ctx)
	}

	receiveFunc := ctx.Clove().Receive(ctx)
	go func() {
		for {
			msg, ok := <- ctx.Messages()
			if !ok {
				ctx.Debug("Clove channel is closed ... stop message-loop")
				break
			}
			switch payload := msg.Payload; payload.(type) {
			case ChildStopped:
				cs := payload.(ChildStopped)
				if cs.Child != nil {
					ctx.RemoveChild(cs.Child)
				}
			case Stop:
				go func() {
					ctx.Debug("PreStop '%v'", ctx.Self().Name())
					if ctx.Clove().PreStop != nil {
						ctx.Clove().PreStop(ctx)
					}

					cl := make([]*CloveRef, len(ctx.Children()))
					copy(cl, ctx.Children())
					for _, child := range cl {
						<- child.Request(Stop{})
						ctx.RemoveChild(child)
					}

					ctx.Debug("PostStop '%v'", ctx.Self().Name())
					if ctx.Clove().PostStop != nil {
						ctx.Clove().PostStop(ctx)
					}

					msg.Reply(Stopped{})
					if parent, ok := ctx.Parent(); ok {
						parent.Tell(ChildStopped{ctx.Self()})
					}

					return
				}()
			default:
				if ctx.Clove().Async {
					go receiveFunc(msg)
				} else {
					receiveFunc(msg)
				}
			}
		}
	}()
	ctx.Debug("PostStart '%v'", ctx.Self().Name())
	if ctx.Clove().PostStart != nil {
		ctx.Clove().PostStart(ctx)
	}
}