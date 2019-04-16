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
	DB  *string
	CSV *CSVConfig
}

type CSVConfig struct {
	Path *string
}

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

	low := strings.ToLower(*conf.DB)
	conf.DB = &low

	switch *conf.DB {
	default:
		log.Fatal("Adapter for ", *conf.DB, " database doesn't exist")
	}

	return &conf
}
