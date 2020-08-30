package golik

type Message struct {
	Payload interface{}
	sender *CloveRef
	reply chan interface{}
	// TODO define message-ids for monitoring and tracking
}

func (m Message) Sender() (*CloveRef, bool) {
	return m.sender, m.sender != nil
}

func (m Message) Reply(result interface{}) {
	m.reply <- result
	close(m.reply)
}

func (m Message) Result() <- chan interface{} {
	return m.reply
}

func NewMessage(sender *CloveRef, payload interface{}) Message {
	return Message{
		Payload: payload,
		reply: make(chan interface{}, 1),
	}
}