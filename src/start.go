package src

import (
	// "net/http"
	"time"
)

var adapter Adapter

type Adapter interface {
	Init() error
	GetCount() (int, error)
	GetCountChan() (chan int, error)
	GetRows(start, end int) ([][]string, error)
	GetRowsDateAverage(start, end time.Time, average int) ([][]string, error)
}

func Start() {
	ReadConfig()
}
