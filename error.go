package golik

type Error struct {
	Message string `json:"error"`
	Code string `json:"code,omitempty"`
	Description string `json:"description,omitempty"`
	Meta map[string]string `json:"meta,omitempty"`
}

func (err *Error) Error() string {
	return err.Message
}

func NewError(err error) *Error {
	return &Error{
		Message: err.Error(),
	}
}
