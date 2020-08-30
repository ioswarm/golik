package crud

import (
	"errors"
	"sync"
	"reflect"

	"github.com/ioswarm/golik"
)

type CreateCommand struct {
	Data interface{}
}

type ReadCommand struct {
	Id interface{}
}

type CRUDCreatePort interface {
	Create(data interface{}, ctx golik.CloveContext) (interface{}, error)
}

type CRUDReadPort interface {
	Read(id interface{}, ctx golik.CloveContext) (interface{}, error)
}

func StatefulCRUD(obj interface{}, config golik.MinionConfig) *golik.Clove {
	return CRUD(obj, golik.MinionConfig{
		Name: config.Name,
		BufferSize: config.BufferSize,
		Async: config.Async,
		Stateful: true,
		Handler: config.Handler,
	})
}

func StatelessCRUD(obj interface{}, config golik.MinionConfig) *golik.Clove {
	return golik.Minion(obj, golik.MinionConfig{
		Name: config.Name,
		BufferSize: config.BufferSize,
		Async: true,
		Stateful: false,
		Handler: config.Handler,
	})
}

func CRUD(obj interface{}, conf golik.MinionConfig) *golik.Clove {
	name := conf.Name
	if obj != nil && name == "" {
		if otype := reflect.TypeOf(obj); otype.Kind() == reflect.Ptr {
			name = otype.Elem().Name()
		} else {
			name = otype.Name()
		}
	}

	crudHandler := conf.Handler
	if crudHandler == nil {
		crudHandler = newCRUDHandler(obj, &conf)	
	}

	return &golik.Clove{
		Name: name,
		Receive: crudHandler.HandleReceive,
		PreStart: func(ctx golik.CloveContext) {
			crudHandler.CallLifeCycle("PreStart", ctx)
		},
		PostStart: func(ctx golik.CloveContext) {
			crudHandler.CallLifeCycle("PostStart", ctx)
		},
		PreStop: func(ctx golik.CloveContext) {
			crudHandler.CallLifeCycle("PreStop", ctx)
		},
		PostStop: func(ctx golik.CloveContext) {
			crudHandler.CallLifeCycle("PostStop", ctx)
		},
		Async: conf.Async,
		BufferSize: conf.BufferSize,
	}
}

func newCRUDHandler(obj interface{}, conf *golik.MinionConfig) *crudHandler {
	return &crudHandler{
		crud: obj,
		conf: conf,
	}
}

type crudHandler struct {
	crud interface{}
	conf *golik.MinionConfig
	mutex sync.Mutex
}

func (ch *crudHandler) CallLifeCycle(methodName string, ctx golik.CloveContext) {
	golik.CallLifeCycle(ch.crud, methodName, ctx)
}

func handleCreateCommand(ccmd CreateCommand, crudminion interface{}, ctx golik.CloveContext) (interface{}, error) {
	switch crudminion.(type) {
	case CRUDCreatePort:
		ccp := crudminion.(CRUDCreatePort)
		return ccp.Create(ccmd.Data, ctx)
	default:
		if result, ok := golik.CallMethod(crudminion, ctx, ccmd); ok {
			switch result.(type) {
			case error:
				return nil, result.(error)
			default:
				return result, nil
			}
		}
	}
	return nil, errors.New("Create method not found")
}

func handleReadCommand(rcmd ReadCommand, crudminion interface{}, ctx golik.CloveContext) (interface{}, error) {
	switch crudminion.(type) {
	case CRUDReadPort:
		crp := crudminion.(CRUDReadPort)
		return crp.Read(rcmd.Id, ctx)
	default:
		if result, ok := golik.CallMethod(crudminion, ctx, rcmd); ok {
			switch result.(type) {
			case error:
				return nil, result.(error)
			default:
				return result, nil
			}
		}
	}
	return nil, errors.New("Read method not found")
}

func (ch *crudHandler) HandleReceive(ctx golik.CloveContext) func(golik.Message) {
	//ctxValue := reflect.ValueOf(ctx)
	return func(msg golik.Message) {
		if ch.crud == nil {
			return
		}
		if ch.conf.Stateful {
			ch.mutex.Lock()
			defer ch.mutex.Unlock()
		}

		msgcontent := msg.Payload
		switch msgcontent.(type) {
		case CreateCommand:
			cc := msgcontent.(CreateCommand)
			res, err := handleCreateCommand(cc, ch.crud, ctx)
			if err != nil {
				// TODO maybe log error
				msg.Reply(err)
				return
			}
			msg.Reply(res)
		case ReadCommand:
			rc := msgcontent.(ReadCommand)
			res, err := handleReadCommand(rc, ch.crud, ctx)
			if err != nil {
				// TODO maybe log error
				msg.Reply(err)
				return
			}
			msg.Reply(res)

		default:
			if result, ok := golik.CallMethod(ch.crud, ctx, msgcontent); ok {
				msg.Reply(result)
			}
		}
	}
}