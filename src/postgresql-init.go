package src

import (
	"database/sql"
	"fmt"

	// "github.com/dgraph-io/badger"
	_ "github.com/lib/pq"
)

// This adapter stores and calculates data in postgresql.

type PSQL struct {
	db             *sql.DB
	dateColumnName string
	columnNames    []string
}

func (p *PSQL) Init() error {
	db, err := sql.Open("postgres", *conf.PSQL.URL)
	if err != nil {
		return err
	}

	p.db = db

	p.columnNames = []string{}

	rows, err := p.db.Query(fmt.Sprint("select column_name from information_schema.columns where table_name = '", escapeString(*conf.PSQL.Table, "'"), "'"))
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			return err
		}
		p.columnNames = append(p.columnNames, columnName)
	}

	p.dateColumnName = p.columnNames[dateColumn]

	if err := rows.Err(); err != nil {
		return err
	}

	if err := p.checkTables(); err != nil {
		return err
	}

	return p.createCache()
}

func CreatePSQL() error {
	adapter = &PSQL{nil, "", nil}
	return adapter.Init()
}

func (p *PSQL) Close() {
	p.db.Close()
}
