package src

func (p *PSQL) GetCount() (int, error) {
	count := 0
	err := p.db.QueryRow("select count(*) from " + *conf.PSQL.Table).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
