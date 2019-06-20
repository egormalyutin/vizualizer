package src

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"time"
)

const randomSymbols string = "qwertyuiopasdfghjklzxcvbnm"
const timeFormat = "2006-01-02 15:04:05"

func interfaceToPointer(i interface{}) *interface{} {
	pointer := reflect.ValueOf(i)
	val := reflect.Indirect(pointer)
	res := val.Interface()
	return &res
}

func escapeString(str, ch string) string {
	return strings.Replace(strings.Replace(str, "\\", "\\\\", -1), ch, "\\"+ch, -1)
}

func randomID() string {
	str := "rid"
	for i := 0; i < 14; i++ {
		id := rand.Intn(len(randomSymbols))
		str += string(randomSymbols[id])
	}
	return str
}

func formatTime(t time.Time) string {
	return t.Format(timeFormat)
}

func mustParseTime(t string) time.Time {
	res, err := time.Parse(timeFormat, t)
	if err != nil {
		panic(err)
	}
	return res
}

func toErr(prefix string, err error) error {
	// return fmt.
}
