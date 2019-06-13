package src

import (
	"fmt"
	"time"
)

func (p *PSQL) GetRows(offset, count int) (chan []Column, chan error, chan bool) {
	chanRes := make(chan []Column)
	chanErr := make(chan error)
	chanSucc := make(chan bool)

	go func() {
		defer close(chanRes)
		defer close(chanErr)
		defer close(chanSucc)

		rows, err := p.db.Query(fmt.Sprint("select * from \"", escapeString(*conf.PSQL.Table, "\""),
			"\" offset ", offset, " limit ", count,
			" order by \"", escapeString(p.dateColumnName, "\""), "\" asc"))

		if err != nil {
			chanErr <- err
		}

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

		chanSucc <- true
	}()

	return chanRes, chanErr, chanSucc
}
