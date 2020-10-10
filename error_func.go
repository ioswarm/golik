package golik

import "fmt"

func (e *Error) Error() string {
	return fmt.Sprintln(e.Code, e.Message)
}

func NewError(message string) *Error {
	return &Error{
		Message: message,
	}
}

func Errorf(format string, a ...interface{}) *Error {
	return NewError(fmt.Sprintf(format, a...))	
}

func Errorln(a ...interface{}) *Error {
	return NewError(fmt.Sprintln(a...))
}
