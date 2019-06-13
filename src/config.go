package src

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	DB     *string
	Format []string
	PSQL   *PSQLConfig `toml:"postgresql"`
}

type PSQLConfig struct {
	URL          *string
	Table        *string
	MinutesTable *string `toml:"minutes-table"`
	HoursTable   *string `toml:"hours-table"`
	DaysTable    *string `toml:"days-table"`
}

var dateColumn = -1

func ReadConfig() *Config {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal("Cannot get current dir: ", err)
	}

	path := filepath.Join(dir, "config.toml")

	if _, err = os.Stat(path); os.IsNotExist(err) {
		log.Fatal("Not found config at ", path)
	}

	log.Print("Found config at ", path)

	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal("Cannot read config: ", err)
	}

	var conf Config

	if _, err = toml.Decode(string(data), &conf); err != nil {
		log.Fatal("Error while parsing config: ", err)
	}

	// VALIDATE

	if conf.DB == nil {
		log.Fatal("No database specified in config")
	}

	if conf.Format == nil {
		log.Fatal("No format specified in config")
	}

	for i, format := range conf.Format {
		conf.Format[i] = strings.ToLower(format)
		format = conf.Format[i]

		switch format {
		case "date":
			if dateColumn != -1 {
				log.Fatal("There can be only one date column")
			}
			dateColumn = i

		case "number":

		default:
			log.Fatal("Format \"", format, "\" not exists")
		}
	}

	if dateColumn == -1 {
		log.Fatal("No data column in format")
	}

	low := strings.ToLower(*conf.DB)
	conf.DB = &low

	switch *conf.DB {
	case "pq", "pql", "psql", "postgre", "postgres", "postgresql":
		if conf.PSQL == nil {
			log.Fatal("No postgresql settings in config")
		}

		if conf.PSQL.URL == nil {
			log.Fatal("No postgresql URL in config")
		}

		if conf.PSQL.Table == nil {
			log.Fatal("No postgresql table name in config")
		}

		*conf.DB = "postgresql"

	default:
		log.Fatal("Adapter for ", *conf.DB, " database doesn't exist")
	}

	// CACHE TABLES MESSAGE
	if conf.PSQL != nil {
		cacheArr := []string{}

		if conf.PSQL.MinutesTable != nil {
			cacheArr = append(cacheArr, "minutes")
		}
		if conf.PSQL.HoursTable != nil {
			cacheArr = append(cacheArr, "hours")
		}
		if conf.PSQL.DaysTable != nil {
			cacheArr = append(cacheArr, "days")
		}

		switch len(cacheArr) {
		case 0:
			log.Print("WARNING: Not found any PostgreSQL cache tables in config. For greatest performance you must include all of them. You can see instructions for setting up these cache tables in README.")
			// TODO: write about cache tables in README
		case 1:
			log.Print("Found one PostgreSQL cache table in config: " + cacheArr[0] + ".")
			log.Print("TIP: For greatest performance, you can include more PostgreSQL cache tables in config. You can see instructions for setting up these cache tables in README.")
		case 2:
			log.Print("Found two PostgreSQL cache tables in config: " + cacheArr[0] + " and " + cacheArr[1] + ".")
			log.Print("TIP: For greatest performance, you can include more PostgreSQL cache tables in config. You can see instructions for setting up these cache tables in README.")
		case 3:
			log.Print("Found all three PostgreSQL cache tables in config: " + cacheArr[0] + ", " + cacheArr[1] + " and " + cacheArr[2] + ".")
			log.Print("You've got a nice perfomance! :)")
		}
	}

	return &conf
}
