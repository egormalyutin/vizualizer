package src

import (
	"reflect"
)

func interfaceToPointer(i interface{}) *interface{} {
	pointer := reflect.ValueOf(i)
	val := reflect.Indirect(pointer)
	res := val.Interface()
	return &res
}
