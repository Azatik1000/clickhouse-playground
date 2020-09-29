package clickhouse

import (
	"gopkg.in/yaml.v2"
	"log"
	"os"
)

type Version struct {
	Id           string `yaml:"id"`
	ImageVersion string `yaml:"imageVersion"`
}

var Versions []Version

func init() {
	readConfig()
}

func readConfig() {
	// TODO: move config reading to shared place to remove code duplication with templater
	configFile, err := os.OpenFile("config/versions.yaml", os.O_RDONLY, 0)
	if err != nil {
		log.Fatal(err)
	}

	decoder := yaml.NewDecoder(configFile)

	data := struct {
		Versions []Version `yaml:"versions"`
	}{}

	if err := decoder.Decode(&data); err != nil {
		log.Fatal(err)
	}

	Versions = data.Versions
}