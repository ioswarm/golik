package reflect

import (
	"reflect"
)

var (
	errorType = reflect.TypeOf((*error)(nil)).Elem()
)
