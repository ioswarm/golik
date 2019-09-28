package golik

type Subscription struct {
	Clove CloveRef
	Filter func(msg interface{}) bool
}

/*
func pubsub() Producer {
	return func(parent CloveRef, name string) CloveRef {
		c := &clove{
			system: parent.System(),
			parent: parent,
			name: name,
			children: make([]CloveRef, 0),
			messages: make(chan Message, 1000), // TODO buffer-size from settings
			receiver: func(context CloveContext) CloveReceiver{
				return &PubSubReceiver{
					context: context,
					subscriptions: make([]Subscription, 0),
				}
			},
			runnable: DefaultRunnable,
		}

		c.log = newLogrusLogger(map[string]interface{}{
			"name": c.Name(),
			"path": c.Path(),
		})

		c.run()

		return c
	}
}*/

func pubsub() CloveDefinition {
	return CloveDefinition{
		Name: "pubsub",
		Receiver: func(context CloveContext) CloveReceiver {
			return &PubSubReceiver{
				context: context,
				subscriptions: make([]Subscription, 0),
			}
		},
	}
}

type PubSubReceiver struct {
	context CloveContext
	subscriptions []Subscription
}

func (psr *PubSubReceiver) indexOfClove(c CloveRef) int {
	for i, s := range psr.subscriptions {
		if c.Path() == s.Clove.Path() {
			return i
		}
	}
	return -1
}

func (psr *PubSubReceiver) addSubscription(sub Subscription) int {
	if index := psr.indexOfClove(sub.Clove); index > 0 {
		psr.subscriptions[index] = sub
		return index
	}
	psr.subscriptions = append(psr.subscriptions, sub)
	return len(psr.subscriptions)-1
}

func (psr *PubSubReceiver) removeSubscription(c CloveRef) int {
	if index := psr.indexOfClove(c); index >= 0 {
		psr.subscriptions = append(psr.subscriptions[:index], psr.subscriptions[index+1:]...)
	}
	return -1
}

func (psr *PubSubReceiver) Receive(reference CloveRef, messages <-chan Message) {
	go func() {
		defer reference.Debug("Receiver messaging loop ended")
		for {
			msg, ok := <- messages
			if !ok {
				reference.Debug("Receiver channel is closed, no more messages will be processed")
				return
			}
			
			switch payload := msg.Payload(); payload.(type) {
			case SubscribeCommand:
				sub := payload.(SubscribeCommand)
				reference.Debug("Subscribe clove %v", sub.Subscription.Clove.Path())
				psr.addSubscription(sub.Subscription)
			case UnsubscribeCommand:
				usub := payload.(UnsubscribeCommand)
				reference.Debug("Unsubscribe clove %v", usub.Clove.Path())
				psr.removeSubscription(usub.Clove)
			default: 
				for _, sub := range psr.subscriptions {
					if sub.Filter(payload) {
						sub.Clove.Tell(payload)
					}
				}
			} 
		}
	}()
}