package src

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// This adapter calculates everything on server and stores data in PostgresQL.

type PSQL struct {
	db             *sql.DB
	dateColumnName string
}

func (p *PSQL) Init() error {
	db, err := sql.Open("postgres", *conf.PSQL.URL)
	if err != nil {
		return err
	}

	p.db = db

	err = p.db.QueryRow(fmt.Sprint("select column_name from information_schema.columns where table_name = '", *conf.PSQL.Table, "' limit 1 offset ", dateColumn)).Scan(&p.dateColumnName)

	if err != nil {
		return err
	}

	return nil
}

func (p *PSQL) Close() {
	p.db.Close()
}

func (p *PSQL) GetCount() (int, error) {
	count := 0
	err := p.db.QueryRow("select count(*) from " + *conf.PSQL.Table).Scan(&count)
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
		defer close(chanRes)
		defer close(chanErr)
		defer close(chanEnd)

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
					values = append(values, Date{*(pointer.(*time.Time))})
				case "number":
					values = append(values, Number{*(pointer.(*float64))})
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

// func (p *PSQL) GetRowsAvg(start, end time.Time, limit int) (chan []Column, chan error, chan bool, error) {
// whereStr := fmt.Sprint(" where time >= '", start.Format(timeFormat), "' and time <= '", end.Format(timeFormat), "'")
// fmt.Println(whereStr)

// count := 0
// err := p.db.QueryRow("select count(*) from " + *conf.PSQL.Table + whereStr).Scan(&count)
// if err != nil {
// return nil, nil, nil, err
// }

// return nil, nil, nil, nil
// }

func CreatePSQL() error {
	adapter = &PSQL{nil, ""}
	return adapter.Init()
}
