package golik

import (
	"strconv"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
)

var (
	ctxType = reflect.TypeOf((*RouteContext)(nil)).Elem()
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

func handleError(err interface{}) Response {
	switch err.(type) {
	case *Error:
		e := err.(*Error)
		result := Response{
			StatusCode: http.StatusInternalServerError,
			Content: e,
		}
		if val, ok := e.Meta["http.status"]; ok {
			if status, serr := strconv.Atoi(val); serr == nil {
				result.StatusCode = status
			}
		}
		return result
	case error:
		return Response{
			StatusCode: http.StatusInternalServerError,
			Content: Errorln(err.(error)),
		}
	default:
		return Response{
			StatusCode: http.StatusInternalServerError,
			Content: &Error{
				Message: "Unknown error occurd",
			},
		}
	}
}

func handleRoute(ctx RouteContext, f interface{}) Response {
	fType := reflect.TypeOf(f)
	fValue := reflect.ValueOf(f)

	params := make([]reflect.Value, 0)
	if fType.NumIn() == 1 {
		if CompareType(fType.In(0), ctxType) {
			params = append(params, reflect.ValueOf(ctx))
		} else {
			paramValue, err := decodeParam(fType.In(0), ctx.Content())
			if err != nil {
				return Response{
					StatusCode: http.StatusBadRequest,
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
			return Response{
				StatusCode: http.StatusBadRequest,
				Content: err,
			}
		}

		params = append(params, paramValue)
	}

	results := fValue.Call(params)

	if fType.NumOut() == 1 {
		if IsErrorType(fType.Out(0)) {
			if !results[0].IsNil() {
				return handleError(results[0].Interface())
			}
		}
		/*if utils.CompareType(fType.Out(0), reflect.TypeOf(Response{})) {
			return results[0].Interface().(Response)
		}*/
		if resp, ok := results[0].Interface().(Response); ok {
			return resp
		}
		return Response{
			StatusCode: http.StatusOK,
			Content: results[0].Interface(),
		}
	}
	if fType.NumOut() == 2 {
		if !results[1].IsNil() {
			return handleError(results[1].Interface())
		}
		/*if utils.CompareType(fType.Out(0), reflect.TypeOf(Response{})) {
			return results[0].Interface().(Response)
		}*/
		if resp, ok := results[0].Interface().(Response); ok {
			return resp
		}
		return Response{
			StatusCode: http.StatusOK,
			Content: results[0].Interface(),
		}
	}

	return Response{StatusCode: http.StatusOK}
}
