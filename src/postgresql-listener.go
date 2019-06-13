package src

import (
	"time"
)

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

		time.Sleep(time.Second * 5)
	}
}
