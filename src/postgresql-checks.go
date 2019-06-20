package src

import "errors"

func (p *PSQL) checkTables() error {
	check := func(tbl *string, prefix string) error {
		if tbl == nil {
			return nil
		}

		var exists bool
		if err := p.db.QueryRow(`select exists(select datname from pg_catalog.pg_database where datname = '` + escapeString(*tbl, `'`) + `'`).Scan(&exists); err != nil {
			return err
		}
		if !exists {
			return errors.New(prefix + ` "` + escapeString(*tbl, `"`) + `" doesn't exist`)
		}

		return nil
	}

	if err := check(conf.PSQL.Table, "Main table"); err != nil {
		return err
	}
	if err := check(conf.PSQL.MinutesTable, "Minutes cache table"); err != nil {
		return err
	}
	if err := check(conf.PSQL.HoursTable, "Hours cache table"); err != nil {
		return err
	}
	if err := check(conf.PSQL.DaysTable, "Days cache table"); err != nil {
		return err
	}

	return nil
}
