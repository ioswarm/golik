package mongo

import (
	"time"
	"context"
	"reflect"
	"sort"
	"github.com/ioswarm/golik"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	ior "github.com/ioswarm/goutils/reflect"
)

func MongoPool(size int, f func() interface{}) golik.CloveDefinition {
	return golik.WorkerPool(size, mongoDefinition("mongo", f(), NewDefaultSettings()))
}

func MongoPoolOf(name string, size int, f func() interface{}) golik.CloveDefinition {
	return golik.WorkerPool(size, mongoDefinition(name, f(), NewSettings(name)))
}


func Mongo(handler interface{}) golik.CloveDefinition{
	return mongoDefinition("mongo", handler, NewDefaultSettings())
}

func MongoOf(name string, handler interface{}) golik.CloveDefinition {
	return mongoDefinition(name, handler, NewSettings(name))
}

func mongoDefinition(name string, handler interface{}, settings *Settings) golik.CloveDefinition {
	return golik.CloveDefinition{
		Name: name,
		LogParams: map[string]interface{}{
			"uri": settings.URI(),
			"collection": settings.Collection,
		},
		Receiver: func (context golik.CloveContext) golik.CloveReceiver {
			return &MongoReceiver{
				settings: settings,
				MongoHandler: handler,
			}
		},
	}
}


type MongoReceiver struct {
	settings *Settings
	mongoClient *mongo.Client
	MongoHandler interface{}
	ticker *time.Ticker
}

type Ping struct {}

func (r *MongoReceiver) Receive(ref golik.CloveRef, messages <-chan golik.Message) {
	go func() {
		defer ref.Debug("Receiver messaging loop ended")
		for {
			msg, ok := <- messages
			if !ok {
				ref.Debug("Receiver channel is closed, no more messages will be processed")
				return
			}

			switch payload := msg.Payload(); payload.(type) {
			case Ping:
				r.checkConnection(ref)
			default:
				if result, err := callMongoHandlerLogic(&MongoContext{CloveRef: ref, Client: r.mongoClient, Settings: r.settings}, r.MongoHandler, payload); err != nil {
					msg.Reply(err)
				} else if result != nil {
					msg.Reply(result)
				} else {
					ref.Debug("No receiver-result for '%T' reply 'Nothing'", msg.Payload())
					msg.Reply(golik.Nothing())
				}	
			}
		}
	}()
}

func (r *MongoReceiver) checkConnection(ref golik.CloveRef) {
	ref.Debug("Ping database, to check connection")
	ctx, _ := context.WithTimeout(context.Background(), r.settings.PingTimeout)
	err := r.mongoClient.Ping(ctx, readpref.Primary())
	if err != nil {
		ref.Panic("Connection lost from %v: %v", r.settings.URI(), err.Error)
	}
}

func (r *MongoReceiver) PreStart(ref golik.CloveRef) {
	go func() {
		ref.Debug("Connect to ", r.settings.URI())
		ctx, _ := context.WithTimeout(context.Background(), r.settings.ConnectionTimeout)
		client, err := mongo.Connect(ctx, options.Client().ApplyURI(r.settings.URI()))
		if err != nil {
			ref.Panic("Could not connect to %v: %v", r.settings.URI(), err.Error)
		}

		r.mongoClient = client

		r.checkConnection(ref)

		ref.Info("Connection to %v established", r.settings.URI())

		if r.settings.CheckConnectionInterval.Milliseconds() > 0 {
			r.ticker = ref.Ticker(r.settings.CheckConnectionInterval, Ping{})
		}
	}()

	golik.CallLifeCycle(ref, r.MongoHandler, "PreStart")
}

func (r *MongoReceiver) PostStart(ref golik.CloveRef) {
	golik.CallLifeCycle(ref, r.MongoHandler, "PostStart")
}

func (r *MongoReceiver) PreStop(ref golik.CloveRef) {
	golik.CallLifeCycle(ref, r.MongoHandler, "PreStop")
}

func (r *MongoReceiver) PostStop(ref golik.CloveRef) {
	golik.CallLifeCycle(ref, r.MongoHandler, "PostStop")

	if r.ticker != nil {
		r.ticker.Stop()
	}

	ctx := context.Background()
	if err := r.mongoClient.Disconnect(ctx); err != nil {
		ref.Error("Could not disconnection from %v: %v", r.settings.URI(), err.Error)
	}
}

// reflect
func callMongoHandlerLogic(c *MongoContext, handler interface{}, value interface{}) (interface{}, error) {
	if handler != nil {
		mvalue := ior.ToPtrValue(reflect.ValueOf(handler))
		vvalue := reflect.ValueOf(value)
		cvalue := reflect.ValueOf(c)

		meths := ior.FindMethodsOf(mvalue, vvalue.Type(), cvalue.Type())
		if len(meths) > 0 { 
			methodCategory := func(meth reflect.Value) int {
				methType := meth.Type()
				
				if methType.NumOut() == 2 && ior.IsErrorType(methType.Out(1)) {
					return 0
				} else if methType.NumOut() == 2 && ior.IsErrorType(methType.Out(0)) {
					return 2
				} else if methType.NumOut() == 1 && !ior.IsErrorType(methType.Out(0)) {
					return 4
				} else if methType.NumOut() == 1 { 
					return 6
				} else if methType.NumOut() == 0 { 
					return 8
				}
				return 999
			}

			sort.Slice(meths, func(a, b int) bool {
				return methodCategory(meths[a]) < methodCategory(meths[b])
			})

			// TODO ... debug info
			meth := meths[0]
			switch methodCategory(meth) {
			case 0:
				result := meth.Call([]reflect.Value{vvalue, cvalue})
				if err, ok := result[1].Interface().(error); ok {
					return result[0].Interface(), err
				}
				return result[0].Interface(), nil
			case 2:
				result := meth.Call([]reflect.Value{vvalue, cvalue})
				if err, ok := result[0].Interface().(error); ok {
					return result[1].Interface(), err
				}
				return result[1].Interface(), nil
			case 4:
				result := meth.Call([]reflect.Value{vvalue, cvalue})
				return result[0].Interface(), nil
			case 6:
				result := meth.Call([]reflect.Value{vvalue, cvalue})
				if err, ok := result[0].Interface().(error); ok {
					return nil, err
				}
				return nil, nil
			case 8:
				meth.Call([]reflect.Value{vvalue, cvalue})
				return nil, nil
			default:
				// TODO throw warning ... method found but too many or less result-types
			}
		}
	}
	return nil, nil
}