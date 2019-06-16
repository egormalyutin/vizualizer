package src

import (
	"fmt"
	"strings"
	"time"
)

func (p *PSQL) GetRowsAvg(start, end time.Time, count int) (chan []Column, chan error, chan bool) {
	start = start.UTC()
	end = end.UTC()

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

	avCacheLevel := 0
	avCache := *conf.PSQL.Table

	for i := cacheLevel; i >= 1; i-- {
		if i == 3 && conf.PSQL.DaysTable != nil {
			avCacheLevel = 3
			avCache = *conf.PSQL.DaysTable
			break
		} else if i == 2 && conf.PSQL.HoursTable != nil {
			avCacheLevel = 2
			avCache = *conf.PSQL.HoursTable
			break
		} else if i == 1 && conf.PSQL.MinutesTable != nil {
			avCacheLevel = 1
			avCache = *conf.PSQL.MinutesTable
			break
		}
	}

	rid := randomID()
	readTable := escapeString(avCache, "\"")

	var (
		truncStart time.Time
		truncEnd   time.Time
	)

	if avCacheLevel == 3 {
		truncStart = start.Add(24 * time.Hour).Truncate(24 * time.Hour).UTC()
		truncEnd = end.Truncate(24 * time.Hour).UTC()
	} else if avCacheLevel == 2 {
		truncStart = start.Add(time.Hour).Truncate(time.Hour).UTC()
		truncEnd = end.Truncate(time.Hour).UTC()
	} else if avCacheLevel == 3 {
		truncStart = start.Add(time.Minute).Truncate(time.Minute).UTC()
		truncEnd = end.Truncate(time.Minute).UTC()
	} else {
		truncStart = start.UTC()
		truncEnd = end.UTC()
	}

	escDCol := escapeString(p.dateColumnName, "\"")

	query := `
create temp view results_` + rid + ` as
	select *
	from "` + readTable + `"
	where "` + escDCol + `" >= '` + formatTime(truncStart) + `' :: timestamptz at time zone 'UTC'
	  and "` + escDCol + `" <= '` + formatTime(truncEnd) + `' :: timestamptz at time zone 'UTC';

create temp table min_max_` + rid + ` (min, max, diff) as (
	select min, max, max - min as diff from (
		select extract(epoch from min) as min, extract(epoch from max) as max
		from (
			select min("` + escDCol + `") as min, max("` + escDCol + `") as max
			from results_` + rid + `
		) t1
	) t2
);
`

	escapedCols := []string{}
	for _, name := range p.columnNames {
		escapedCols = append(escapedCols, escapeString(name, "\""))
	}

	columnsStr := strings.Join(escapedCols, ", ")

	readOrigTable := escapeString(*conf.PSQL.Table, "\"")

	query += `
create temp table min_orig_` + rid + ` (` + columnsStr + `) as (
	select *
	from "` + readOrigTable + `"
	where "` + escDCol + `" >= '` + formatTime(start) + `' :: timestamp at time zone 'UTC'
	  and "` + escDCol + `" <= '` + formatTime(end) + `' :: timestamp at time zone 'UTC'
	order by "` + escDCol + `" asc
	limit 1
);

create temp table max_orig_` + rid + ` (` + columnsStr + `) as (
	select *
	from "` + readOrigTable + `"
	where "` + escDCol + `" >= '` + formatTime(start) + `' :: timestamp at time zone 'UTC'
	  and "` + escDCol + `" <= '` + formatTime(end) + `' :: timestamp at time zone 'UTC'
	order by "` + escDCol + `" desc
	limit 1
);
`

	query += `
select * from (
	(select * from min_orig_` + rid + `)
	union
	(select
`

	sStrings := []string{}

	for _, name := range escapedCols {
		if name == escDCol {
			sStrings = append(sStrings, `to_timestamp(avg(extract(epoch from "`+escDCol+`"))) as "`+escDCol+`"`)
		} else {
			sStrings = append(sStrings, `avg("`+name+`") as "`+name+`"`)
		}
	}

	query += strings.Join(sStrings, ",\n")

	query += `
	from results_` + rid + `
	group by floor((extract(epoch from "` + escDCol + `") - (select min from min_max_` + rid + `)) / (select diff from min_max_` + rid + `) * ` + fmt.Sprint(count-3) + `))
	union
	(select * from max_orig_` + rid + `)
) ret1
order by "` + escDCol + `" asc;

drop view if exists results_` + rid + `;
drop table if exists min_max_` + rid + `;
drop table if exists min_orig_` + rid + `;
drop table if exists max_orig_` + rid + `;
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
