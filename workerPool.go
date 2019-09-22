package golik

import (
	"fmt"
	"strconv"
	"sort"
)

func WorkerPool(size int, worker Producer) Producer {
	return func(parent CloveRef, name string) CloveRef {
		c := &clove{
			system: parent.System(),
			parent: parent,
			name: name,
			children: make([]CloveRef, 0),
			messages: make(chan Message, 1000), // TODO buffer-size from settings
			receiver: func(context CloveContext) CloveReceiver{
				return &WorkerPipeReceiver{
					context: context,
				}
			},
			runnable: defaultRunnable,
		}

		sz := size
		if sz <= 0 {
			sz = 1
		}

		c.log = newLogrusLogger(map[string]interface{}{
			"name": c.Name(),
			"path": c.Path(),
			"pool-size": strconv.Itoa(sz),
		})

		c.run()

		for i := 0; i < sz; i++ {
			c.Of(worker, fmt.Sprintf("%v_%x", name, i))
		}

		return c
	}
}

type WorkerPipeReceiver struct {
	context CloveContext
}

type chanSize struct {
	size int
	path string
	channel chan<- Message
}

func (cs *chanSize) string() string {
	return fmt.Sprintf("%v of size %v", cs.path, cs.size)
}

func (r *WorkerPipeReceiver) Receive(reference CloveRef, messages <-chan Message) {
	go func() {
		defer reference.Debug("Receiver messaging loop ended")
		for {
			msg, ok := <- messages
			if !ok {
				reference.Debug("Receiver channel is closed, no more messages will be processed")
				return
			}
			
			sizes := make([]chanSize, len(reference.Children()))
			for i, child := range reference.Children() {
				size := chanSize{size: child.ChannelSize(), path: child.Path(), channel: child.Channel()}
				//reference.Debug("Measure channel-size of %v = %v", size.path, size.size)
				sizes[i] = size
			}
			sort.Slice(sizes, func(i, j int) bool{
				return sizes[i].size < sizes[j].size
			})

			reference.Debug("Send message to %v", sizes[0].string())
			sizes[0].channel <- msg
		}
	}()
}