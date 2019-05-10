package src

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// This adapter calculates everything on server and stores data in PostgresQL.

type PSQL struct {
	db *sql.DB
}

func (p *PSQL) Init() error {
	db, err := sql.Open("postgres", *conf.PSQL.URL)
	p.db = db
	return err
}

func (p *PSQL) Close() {
	p.db.Close()
}

func (p *PSQL) GetCount() (int, error) {
	rows, err := p.db.Query("select count(*) from " + *conf.PSQL.Table)
	if err != nil {
		return 0, err
	}

	defer rows.Close()
	rows.Next()

	count := 0

	err = rows.Scan(&count)
	if err != nil {
		return 0, err
	}

	err = rows.Err()
	if err != nil {
		return 0, err
	}

	return count, nil
}

func (p *PSQL) RunListener() error {
	count := -1

	for {
		newCount, err := p.GetCount()
		if err != nil {
			return err
		}

		if newCount != count {
			count = newCount
			bus.Publish("count", count)
		}

		time.Sleep(time.Second * 2)
	}
}

func (p *PSQL) GetRows(offset, count int) (chan []Column, chan error, chan bool, error) {
	rows, err := p.db.Query(fmt.Sprint("select * from ", *conf.PSQL.Table, " offset ", offset, " limit ", count))
	if err != nil {
		return nil, nil, nil, err
	}

	chanRes := make(chan []Column)
	chanErr := make(chan error)
	chanEnd := make(chan bool)

	go func() {
		defer rows.Close()
		for rows.Next() {
			pointers := []interface{}{}

			for _, format := range conf.Format {
				switch format {
				case "date":
					var date time.Time
					pointers = append(pointers, &date)

				case "number":
					var number float64
					pointers = append(pointers, &number)
				}
			}

			err = rows.Scan(pointers...)
			if err != nil {
				chanErr <- err
				return
			}

			values := []Column{}

			for i, pointer := range pointers {
				switch conf.Format[i] {
				case "date":
					values = append(values, Date{date: *(pointer.(*time.Time))})
				case "number":
					values = append(values, Number{number: *(pointer.(*float64))})
				}
			}

			chanRes <- values
		}

		err = rows.Err()
		if err != nil {
			chanErr <- err
		}

		chanEnd <- true
	}()

	return chanRes, chanErr, chanEnd, nil
}

func CreatePSQL() error {
	adapter = &PSQL{nil}
	return adapter.Init()
}
