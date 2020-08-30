package http

import (
	"encoding/json"
	"errors"
	"io"
	ht "net/http"
	"reflect"

	"github.com/ioswarm/golik"
	"github.com/ioswarm/golik/utils"
)

var (
	ctxType = reflect.TypeOf((*golik.RouteContext)(nil)).Elem()
)

func validateRouteFunc(f interface{}) error {
	if f == nil {
		return errors.New("Route.Handle is nil")
	}
	// TODO implement validation
	return nil
}

func createInstancePtrOf(t reflect.Type) reflect.Value {
	if t.Kind() == reflect.Ptr {
		return reflect.New(t.Elem())
	}
	return reflect.New(t)
}

func decodeParam(inType reflect.Type, reader io.Reader) (reflect.Value, error) {
	inPtrValue := createInstancePtrOf(inType)
	if err := json.NewDecoder(reader).Decode(inPtrValue.Interface()); err != nil {
		return reflect.ValueOf(nil), err
	}
	if inType.Kind() == reflect.Ptr {
		return inPtrValue, nil
	} else {
		return reflect.Indirect(inPtrValue), nil
	}
}

func handleRoute(ctx golik.RouteContext, f interface{}) golik.Response {
	fType := reflect.TypeOf(f)
	fValue := reflect.ValueOf(f)

	params := make([]reflect.Value, 0)
	if fType.NumIn() == 1 {
		if utils.CompareType(fType.In(0), ctxType) {
			params = append(params, reflect.ValueOf(ctx))
		} else {
			paramValue, err := decodeParam(fType.In(0), ctx.Content())
			if err != nil {
				return golik.Response{
					StatusCode: ht.StatusBadRequest,
					Content: err,
				}
			}

			params = append(params, paramValue)
		}
	}
	if fType.NumIn() == 2 {
		params = append(params, reflect.ValueOf(ctx))
		paramValue, err := decodeParam(fType.In(1), ctx.Content())
		if err != nil {
			return golik.Response{
				StatusCode: ht.StatusBadRequest,
				Content: err,
			}
		}

		params = append(params, paramValue)
	}

	results := fValue.Call(params)

	if fType.NumOut() == 1 {
		if utils.IsErrorType(fType.Out(0)) {
			err := results[0].Interface().(error)
			return golik.Response{
				StatusCode: ht.StatusInternalServerError,
				Content: err,
			}
		}
		if utils.CompareType(fType.Out(0), reflect.TypeOf(golik.Response{})) {
			return results[0].Interface().(golik.Response)
		}
		return golik.Response{
			StatusCode: ht.StatusOK,
			Content: results[0].Interface(),
		}
	}
	if fType.NumOut() == 2 {
		err := results[1].Interface().(error)
		if err != nil {
			return golik.Response{
				StatusCode: ht.StatusInternalServerError,
				Content: err,
			}
		}
		if utils.CompareType(fType.Out(0), reflect.TypeOf(golik.Response{})) {
			return results[0].Interface().(golik.Response)
		}
		return golik.Response{
			StatusCode: ht.StatusOK,
			Content: results[0].Interface(),
		}
	}

	return golik.Response{StatusCode: ht.StatusOK}
}
