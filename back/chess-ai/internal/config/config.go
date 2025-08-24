package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Http `yaml:"server"`
	Database `yaml:"database"`
}

func MustLoad() (*Config, error) {
	var cfg Config
	err := cleanenv.ReadConfig("./.yaml", &cfg)
	if 
}