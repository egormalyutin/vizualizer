package src

import (
	"fmt"
	"log"
	"time"
)

func (p *PSQL) GetRowsAvg(start, end time.Time, count int) (chan []Column, chan error, chan bool) {
	timeDiff := end.Sub(start)

	cacheLevel := 0
	ti := timeDiff / time.Duration(count)

	if ti >= (24 * time.Hour) {
		cacheLevel = 3
	} else if ti >= time.Hour {
		cacheLevel = 2
	} else if ti >= time.Minute {
		cacheLevel = 1
	}

	avCache := *conf.PSQL.Table

	for i := cacheLevel; i >= 1; i-- {
		log.Print(conf.PSQL)
		if i == 3 && conf.PSQL.DaysTable != nil {
			avCache = *conf.PSQL.DaysTable
			break
		} else if i == 2 && conf.PSQL.HoursTable != nil {
			avCache = *conf.PSQL.HoursTable
			break
		} else if i == 1 && conf.PSQL.MinutesTable != nil {
			avCache = *conf.PSQL.MinutesTable
			break
		}
	}

	rid := randomID()
	readTable := escapeString(avCache, "\"")
	log.Print(readTable)

	query := `
drop view if exists results_` + rid + `;
drop table if exists min_max_` + rid + `;

create temp view results_` + rid + ` as
	select *
	from "` + readTable + `"
	where time >= '` + formatTime(start) + `' 
	  and time <= '` + formatTime(end) + `';

create temp table min_max_` + rid + ` (min, max, diff) as (
	select min, max, max - min as diff from (
		select extract(epoch from min) as min, extract(epoch from max) as max
		from (
			select min(time) as min, max(time) as max
			from results_` + rid + `
		) t1
	) t2
);

select
to_timestamp(avg(extract(epoch from time))) as time,
avg(voltage) as voltage,
avg(amperage) as amperage,
avg(power) as power,
avg(energy_supplied) as energy_supplied,
avg(energy_received) as energy_received
from results_` + rid + `
group by floor((extract(epoch from time) - (select min from min_max_` + rid + `)) / (select diff from min_max_` + rid + `) * ` + fmt.Sprint(count-1) + `);

drop view if exists results_` + rid + `;
drop table if exists min_max_` + rid + `;
`

	chanRes := make(chan []Column)
	chanErr := make(chan error)
	chanSucc := make(chan bool)

	go func() {
		defer close(chanRes)
		defer close(chanErr)
		defer close(chanSucc)

		rows, err := p.db.Query(query)

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
