package golik

import (
	"fmt"
	"strconv"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func (e *Error) Error() string {
	return fmt.Sprintln(e.Code, e.Message)
}

func (e *Error) WithCode(code string) *Error {
	e.Code = code
	return e
}

func (e *Error) WithDescription(description string) *Error {
	e.Description = description
	return e
}

func (e *Error) WithMeta(key string, value string) *Error {
	if e.Meta == nil {
		e.Meta = make(map[string]string)
	}
	e.Meta[key] = value
	return e
}

func (e *Error) MetaValue(key string) (string, bool) {
	value, ok := e.Meta[key]
	return value, ok
}

func (e *Error) WithHttpStatus(status int) *Error {
	return e.WithMeta("http.status", strconv.Itoa(status))
}

func (e *Error) HttpStatus() int {
	if value, ok := e.MetaValue("http.status"); ok {
		if status, err := strconv.Atoi(value); err == nil {
			return status
		}
		return 500
	}
	return 200
}




func ErrorOf(i interface{}) *Error {
	if i == nil {
		return NewError("Nil pointer")
	}
	switch i.(type) {
	case *Error:
		return i.(*Error)
	case Error:
		err := i.(Error)
		return &err
	case error:
		err := i.(error)
		return NewError(err.Error())
	default:
		return Errorf("%v", i)
	}
}

func NewError(message string) *Error {
	return &Error{
		Message: message,
		Timestamp: timestamppb.Now(),
		Meta: make(map[string]string),
	}
}

func Errorf(format string, a ...interface{}) *Error {
	return NewError(fmt.Sprintf(format, a...))	
}

func Errorln(a ...interface{}) *Error {
	return NewError(fmt.Sprintln(a...))
}
