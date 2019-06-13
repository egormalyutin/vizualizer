package src

import (
	"log"
	"strconv"
	"time"

	eventBus "github.com/asaskevich/EventBus"
)

////// GLOBAL VARS //////
var (
	// DB adapter
	adapter Adapter

	conf *Config

	// Main event bus.
	bus = eventBus.New()

	// Is new rows listener already started?
	listenerStarted = false
)

////// COLUMN TYPES //////
type Date struct {
	Date time.Time
}

func (d Date) ToCSV() string {
	return d.Date.UTC().Format(timeFormat)
}

type Number struct {
	Number float64
}

func (n Number) ToCSV() string {
	return string(strconv.FormatFloat(n.Number, 'f', -1, 64))
}

////// MAIN TYPES //////

// Column can be Date or Number
type Column interface {
	// Transform into CSV string
	ToCSV() string
}

type Adapter interface {
	Init() error
	Close()

	// Return count of rows.
	GetCount() (int, error)

	// Run rows listener. Listener must publish event "count" with new count to main event bus. Call only once.
	// TODO: last row
	// TODO: cache days
	RunListener() error

	// Chan is here to not swamp RAM.
	GetRows(offset, count int) (chan []Column, chan error, chan bool)

	// Gets rows in range of given dates and averages them.
	GetRowsAvg(start, end time.Time, count int) (chan []Column, chan error, chan bool)
}

func Start() {
	conf = ReadConfig()

	switch *conf.DB {
	case "postgresql":
		err := CreatePSQL()
		if err != nil {
			log.Fatal(err)
		}
	}

	defer adapter.Close()

	cres, cerr, cend := adapter.GetRowsAvg(mustParseTime("2015-11-24 05:35:27"), mustParseTime("2016-03-01 06:36:12"), 100)

For1:
	for {
		select {
		case line := <-cres:
			log.Print(line)
		case err := <-cerr:
			log.Print(err)
		case <-cend:
			log.Print("SUCC")
			break For1
		}
	}

	log.Fatal(adapter.RunListener())
}
