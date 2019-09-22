package golik

type Message interface {
	Payload() interface{}
	Result() <-chan interface{}
	Reply(result interface{})
}

type DefaultMessage struct {
	payload interface{}
	resultChannel chan interface{}
}

func NewMessage(payload interface{}) Message {
	return &DefaultMessage{
		payload: payload,
		resultChannel: make(chan interface{}, 1),
	}
}

func (dm *DefaultMessage) Payload() interface{} {
	return dm.payload
}

func (dm *DefaultMessage) Result() <-chan interface{} {
	return dm.resultChannel
}

func (dm *DefaultMessage) Reply(result interface{}) {
	dm.resultChannel <- result
	close(dm.resultChannel)
}