package config

import "github.com/ilyakaznacheev/cleanenv"

type Config struct {
	Host           string `yaml:"host"`
	Port           string `yaml:"port"`
	ComputingPOWER int    `yaml:"computing_power"`
}

func NewConfig() *Config {
	var cfg Config

	if err := cleanenv.ReadConfig("./config/config.yml", &cfg); err != nil {
		return nil
	}

	return &cfg
}
