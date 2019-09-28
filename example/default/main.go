package main

import (
	"strconv"
	"fmt"
	"time"
	"github.com/ioswarm/golik"
)

type TestM struct{}

func (t *TestM) PostStart(c golik.CloveRef) {
	c.Info("TestM started")
}

func (*TestM) PostStop(c golik.CloveRef) {
	c.Info("TestM stopped")
}

func (*TestM) HandleInt(i int) int {
	return i * i
}

func (*TestM) LogString(s string, c golik.CloveRef) {
	c.Info("Receive: %v", s)
}

func (t *TestM) HandleIntRoute() *golik.Route {
	return golik.GET("/test/{num}").Handle( func(ctx *golik.RouteContext) golik.Response { 
		if val, ok := ctx.Params()["num"]; ok {
			if i, err := strconv.Atoi(val); err == nil {
				return golik.OK(t.HandleInt(i))
			} else {
				return golik.BadRequest(err.Error())
			}
		}
		return golik.InternalServerError("")
	})
}

func (t *TestM) HandleIntCloveRoute() golik.CloveRoute {
	return func (c golik.CloveRef) *golik.Route {
		return golik.GET("/test2/{num}").Handle( func(ctx *golik.RouteContext) golik.Response {
			if val, ok := ctx.Params()["num"]; ok {
				if i, err := strconv.Atoi(val); err == nil {
					switch res := <- c.Ask(i, time.Second); res.(type) {
					case int:
						return golik.OK(res)
					case error:
						return golik.InternalServerError(res.(error).Error())
					default:
						return golik.NoContent("")
					}
				} else {
					return golik.BadRequest(err.Error)
				}
			}
			return golik.InternalServerError("Num not set")
		})
	}
}

// TODO explain in more detail
func main() {
	system := golik.GolikSystem()

	
	ref := system.Of(golik.CloveDefinition{
		Name: "TestA",
		Receiver: golik.Minion(&TestM{}),
	})

	ref.Tell("Hallo Welt")
	
	fmt.Printf("Result for int %v = %v\n", 155, <- ref.Ask(155, time.Second))

	<-system.Terminated()
}