package golik

import (
	"time"
	"errors"
)

func await(message interface{}, channel chan<- Message, timeout time.Duration) <-chan interface{} {
	result := make(chan interface{}, 1)
	msg := NewMessage(message)
	go func() {
		channel <- msg
		select {
		case res := <-msg.Result():
			result <- res
			close(result)
		case <-time.After(timeout):
			result <- errors.New("Timeout") // TODO message
			close(result)
		}
	}()
	return result
}

/* runnable functions */

func defaultRunnable(c *clove) {
	receiver := c.receiver(c)
	CallLifeCycle(c, receiver, "PreStart")

	receiverChannel := make(chan Message)
	receiver.Receive(c, receiverChannel)

	go func(){
		defer c.Debug("Messaging loop ended")
		for {
			msg, ok := <- c.messages
			if !ok {
				c.Warn("Messages channel is closed, no more messages will be processed")
				break
			}
			c.Debug("Receive message of %T", msg.Payload())
			switch payload := msg.Payload(); payload.(type) {
			case StopCommand:
				cmd := payload.(StopCommand)
				CallLifeCycle(c, receiver, "PreStop")
				
				for _, child := range c.Children() {
					// TODO handle result and remove ref from children
					switch res := <- await(parentalStop(c), child.Channel(), 1 * time.Second); res.(type) { // TODO configure timeout
					case error:
						c.Warn("%v while sending StopCommand to child %v", res, child.Name())
					default:
						// do nothing	
					}
					c.removeChild(child)
				}

				// Sending Stop to receiver
				switch res := <- await(Stop(), receiverChannel, 1 * time.Second); res.(type) { // TODO configure timeout
				case error:
					c.Warn("%v while sending StopCommand to receiver", res)
				default:
					// do nothing
				}

				// close channels
				close(receiverChannel)
				close(c.messages)

				// remove parent
				if c.Parent() != nil && cmd.sender == nil {
					c.Parent().Tell(ChildStopped(c))
				}
				c.parent = nil
				CallLifeCycle(c, receiver, "PostStop")

				msg.Reply(Stopped())
				return
			case ChildStoppedEvent:
				cmd := payload.(ChildStoppedEvent)
				if cmd.sender != nil {
					c.removeChild(cmd.sender)
				}
				msg.Reply(Nothing())
			default: 
				receiverChannel <- msg
			}
		}
	}()

	CallLifeCycle(c, receiver, "PostStart")
}