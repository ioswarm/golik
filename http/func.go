package http

import (
	"strconv"
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

func handleError(err interface{}) golik.Response {
	switch err.(type) {
	case error:
		return golik.Response{
			StatusCode: ht.StatusInternalServerError,
			Content: golik.NewError(err.(error)),
		}
	case *golik.Error:
		e := err.(*golik.Error)
		result := golik.Response{
			StatusCode: ht.StatusInternalServerError,
			Content: e,
		}
		if val, ok := e.Meta["http.status"]; ok {
			if status, serr := strconv.Atoi(val); serr == nil {
				result.StatusCode = status
			}
		}
		return result
	default:
		return golik.Response{
			StatusCode: ht.StatusInternalServerError,
			Content: &golik.Error{
				Message: "Unknown error occurd",
			},
		}
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
			if !results[0].IsNil() {
				return handleError(results[0].Interface())
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
		if !results[1].IsNil() {
			return handleError(results[1].Interface())
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
