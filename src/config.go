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
	PSQL   *PSQLConfig `toml:"pq,pql,psql,postgres,postgresql"`
}

type PSQLConfig struct {
	URL   *string
	Table *string
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

	return &conf
}
