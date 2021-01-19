package golik

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

var (
	ctxType = reflect.TypeOf((*RouteContext)(nil)).Elem()
	msgType = reflect.TypeOf((*proto.Message)(nil)).Elem()
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

	result := inPtrValue.Interface()
	switch result.(type) {
	case proto.Message:
		pmsg := result.(proto.Message)
		if err := jsonpb.Unmarshal(reader, pmsg); err != nil {
			return reflect.ValueOf(nil), err
		}
	default:
		if err := json.NewDecoder(reader).Decode(result); err != nil {
			return reflect.ValueOf(nil), err
		}
	}

	if inType.Kind() == reflect.Ptr {
		return inPtrValue, nil
	}
	return reflect.Indirect(inPtrValue), nil
}

func handleError(err interface{}) Response {
	switch err.(type) {
	case *Error:
		e := err.(*Error)
		result := Response{
			StatusCode: http.StatusInternalServerError,
			Content:    e,
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
			Content:    Errorln(err.(error)),
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
					Content:    err,
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
				Content:    err,
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
			Content:    results[0].Interface(),
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
			Content:    results[0].Interface(),
		}
	}

	return Response{StatusCode: http.StatusOK}
}

func defaultResponseWriter(ctx HttpRouteContext, w http.ResponseWriter, resp *Response) {
	//w.Header().Set("Content-Type", "application/json; utf-8")
	for key := range resp.Header {
		w.Header().Set(key, resp.Header.Get(key))
	}

	if resp.Content != nil {
		switch resp.Content.(type) {
		case []byte:
			if w.Header().Get("Content-Type") == "" {
				w.Header().Set("Content-Type", "application/octet-stream")
			}
			w.WriteHeader(resp.StatusCode)

			buf := resp.Content.([]byte)
			w.Write(buf)
		case []proto.Message:
			if w.Header().Get("Content-Type") == "" {
				w.Header().Set("Content-Type", "application/json; utf-8")
			}
			w.WriteHeader(resp.StatusCode)

			ps := resp.Content.([]proto.Message)
			m := jsonpb.Marshaler{}
			slist := make([]string, 0)
			for _, msg := range ps {
				s, err := m.MarshalToString(msg)
				if err != nil {
					ctx.Warn(err.Error())
					continue
				}
				slist = append(slist, s)
			}
			result := "[" + strings.Join(slist, ",") + "]"
			w.Write([]byte(result))
		case proto.Message:
			if w.Header().Get("Content-Type") == "" {
				w.Header().Set("Content-Type", "application/json; utf-8")
			}
			w.WriteHeader(resp.StatusCode)

			pm := resp.Content.(proto.Message)
			m := jsonpb.Marshaler{}
			if err := m.Marshal(w, pm); err != nil {
				ctx.Warn(err.Error())
			}
		default:
			if w.Header().Get("Content-Type") == "" {
				w.Header().Set("Content-Type", "application/json; utf-8")
			}
			w.WriteHeader(resp.StatusCode)

			if err := json.NewEncoder(w).Encode(resp.Content); err != nil {
				ctx.Warn(err.Error())
			}
		}
	}
}
