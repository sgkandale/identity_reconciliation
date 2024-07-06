package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Server   ServerConfig   `yaml:"server" env:"server"`
	Database DatabaseConfig `yaml:"database" env:"database"`
}

type DatabaseConfig struct {
	Type    string `yaml:"type" env:"type"`
	Uri     string `yaml:"uri" env:"uri"`
	Timeout int    `yaml:"timeout" env:"timeout"`
}

type ServerConfig struct {
	Port int `yaml:"port" env:"port"`
}

const configFile = "config.yaml"

func ReadConfig() *Config {
	var cfg Config
	err := cleanenv.ReadConfig(configFile, &cfg)
	if err != nil {
		log.Fatalf("[ERROR] reading config file %s : %s", configFile, err.Error())
	}
	return &cfg
}
