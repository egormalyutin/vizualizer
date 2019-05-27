package src

import (
	"log"
	"strconv"
	"time"

	eventBus "github.com/asaskevich/EventBus"
)

const timeFormat = "2006-01-02 15:04:05"

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
	return d.Date.Format(timeFormat)
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
	GetRows(offset, count int) (chan []Column, chan error, chan bool, error)

	// Gets rows in range of given dates and averages them.
	// GetRowsAvg(start, end time.Time, period string) (chan []Column, chan error, chan bool, error)
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

	log.Fatal(adapter.RunListener())
}
