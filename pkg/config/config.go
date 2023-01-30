package config

import (
	"log"

	"github.com/spf13/viper"
)

type Destination struct {
	Namespace string
	Name      string
}

type Secret struct {
	Namespace    string
	Name         string
	Destinations []Destination
}

type Config struct {
	Secrets []Secret
}

func GetConfig() *Config {
	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		log.Fatal(err.Error())
	}
	return cfg
}
