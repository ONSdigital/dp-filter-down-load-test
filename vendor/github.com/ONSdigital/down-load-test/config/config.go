package config

import (
	"io/ioutil"
	"gopkg.in/yaml.v2"
	"os"
	"github.com/ONSdigital/go-ns/log"
)

type Model struct {
	FilterAPIHost string   `yaml:"filter-api-host"`
	Filters       []string `yaml:"filters"`
	InstanceID    string   `yaml:"instanceID"`
}

func Load() Model {
	source, err := ioutil.ReadFile("config.yml")
	if err != nil {
		log.Error(err, nil)
		os.Exit(1)
	}

	var config Model
	if err := yaml.Unmarshal(source, &config); err != nil {
		log.ErrorC("failed to load config.yml", err, nil)
		os.Exit(1)
	}
	log.Info("config", log.Data{"": config})
	return config
}
