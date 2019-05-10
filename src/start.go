package src

import (
	"log"
	"strconv"
	"strings"
	"time"

	eventBus "github.com/asaskevich/EventBus"
)

var (
	// DB adapter
	adapter Adapter

	conf *Config

	// Main event bus.
	bus = eventBus.New()

	// Is new rows listener already started?
	listenerStarted = false
)

type Date struct {
	date time.Time
}

func (d Date) ToCSV() string {
	return d.date.Format("2006-01-02 15:04:05")
}

type Number struct {
	number float64
}

func (n Number) ToCSV() string {
	return string(strconv.FormatFloat(n.number, 'f', -1, 64))
}

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
	RunListener() error

	// Chan is here to not swamp RAM.
	GetRows(offset, count int) (chan []Column, chan error, chan bool, error)

	// Gets rows in range of given dates and averages them.
	// GetRows(start, end time.Time, count int) (chan []Column, chan error, chan bool, error)
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

	bus.Subscribe("count", log.Print)

	resChan, errChan, endChan, err := adapter.GetRows(0, 300)
	if err != nil {
		log.Fatal(err)
	}

Handler:
	for {
		select {
		case row := <-resChan:
			acc := []string{}

			for _, item := range row {
				acc = append(acc, item.ToCSV())
			}

			log.Print(strings.Join(acc, " "))

		case err := <-errChan:
			log.Fatal(err)
		case <-endChan:
			break Handler
		}
	}

	log.Fatal(adapter.RunListener())
}
