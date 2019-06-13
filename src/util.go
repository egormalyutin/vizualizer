package src

import (
	"reflect"
	"strings"
)

func interfaceToPointer(i interface{}) *interface{} {
	pointer := reflect.ValueOf(i)
	val := reflect.Indirect(pointer)
	res := val.Interface()
	return &res
}

func escapeString(str, ch string) string {
	return strings.Replace(str, ch, "\\"+ch, -1)
}
