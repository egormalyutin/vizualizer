package src

import (
	"log"
	"strconv"
	"time"

	eventBus "github.com/asaskevich/EventBus"
)

const timeFormat = "2006-01-02 15:04:05"

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
	// GetRowsAvg(start, end time.Time, count int) (chan []Column, chan error, chan bool, error)
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

	// bus.Subscribe("count", log.Print)

	strs := [][]string{
		{"2018-02-01 00:00:00", "10", "20", "30", "40", "50"},
		{"2018-02-01 00:00:05", "20", "20", "30", "40", "50"},
		{"2018-02-01 00:00:10", "30", "20", "30", "40", "50"},
		{"2018-02-01 00:00:15", "40", "20", "30", "40", "50"},
	}

	rows := [][]Column{}

	for _, line := range strs {
		row := []Column{}

		for i, tp := range conf.Format {
			switch tp {
			case "date":
				t, err := time.Parse(timeFormat, line[i])
				if err != nil {
					log.Fatal(err)
				}
				row = append(row, Date{t})

			case "number":
				n, err := strconv.ParseFloat(line[i], 64)
				if err != nil {
					log.Fatal(err)
				}
				row = append(row, Number{n})

			}
		}

		rows = append(rows, row)
	}

	log.Print(rows)

	ch := make(chan []Column)

	go func() {
		for _, row := range rows {
			ch <- row
		}
		close(ch)
	}()

	res := interpolateChan(ch, rows[0], rows[len(rows)-1], 3)

	i := 0
	for {
		r, more := <-res
		if !more {
			log.Print("end")
			return
		}

		log.Print("Line ", i, ": ", r)
		i++
	}

	// _, _, _, err = adapter.GetRowsAvg(start, end, 20)
	// if err != nil {
	// log.Fatal(err)
	// }

	// Handler:
	// for {
	// select {
	// case row := <-resChan:
	// acc := []string{}

	// for _, item := range row {
	// acc = append(acc, item.ToCSV())
	// }

	// log.Print(strings.Join(acc, " "))

	// case err := <-errChan:
	// log.Fatal(err)
	// case <-endChan:
	// break Handler
	// }
	// }

	log.Fatal(adapter.RunListener())
}
