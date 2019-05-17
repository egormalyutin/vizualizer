package src

import (
	"reflect"
	"time"
)

func interfaceToPointer(i interface{}) *interface{} {
	pointer := reflect.ValueOf(i)
	val := reflect.Indirect(pointer)
	res := val.Interface()
	return &res
}

// Chan can be closed
func interpolateChan(ch chan []Column, firstSearchRow []Column, lastSearchRow []Column, count int) chan []Column {
	firstSearchDateInt := firstSearchRow[dateColumn].(Date).Date.Unix()
	lastSearchDateInt := lastSearchRow[dateColumn].(Date).Date.Unix()

	x := firstSearchDateInt
	step := (lastSearchDateInt - firstSearchDateInt) / int64(count)

	res := make(chan []Column)

	var (
		lastRow     []Column
		lastDateInt int64
	)

	go func() {
		defer close(res)

		for {
			row, more := <-ch
			if !more {
				return
			}

			var dateInt int64

			if lastRow == nil {
				lastRow = row
				lastDateInt := lastRow[dateColumn].(Date).Date.Unix()
				dateInt = lastDateInt
			} else {
				dateInt = row[dateColumn].(Date).Date.Unix()
			}

			if lastDateInt <= x && x <= dateInt {
				// right row!
				k := float64(x-firstSearchDateInt) / float64(dateInt-lastDateInt)
				ret := []Column{}
				newDate := Date{time.Unix(x, 0)}
				for i, tp := range conf.Format {
					switch tp {
					case "date":
						ret = append(ret, newDate)
					case "number":
						lastNum := row[i].(Number).Number
						currNum := lastRow[i].(Number).Number
						resNum := lastNum + (k * (currNum - lastNum))
						ret = append(ret, Number{resNum})
					}
				}
				res <- ret
				x += step
			}

			lastRow = row
			lastDateInt = dateInt
		}
	}()

	return res
}
