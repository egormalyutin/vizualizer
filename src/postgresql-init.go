package src

import (
	"database/sql"
	"fmt"

	// "github.com/dgraph-io/badger"
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

func CreatePSQL() error {
	adapter = &PSQL{nil, ""}
	return adapter.Init()
}

func (p *PSQL) Close() {
	p.db.Close()
}
