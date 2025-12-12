package config

import (
	"sync"

	"github.com/caarlos0/env/v11"
)

const DEV = "dev"
const PROD = "prod"

type Config struct {
	Env              string `env:"ENV" default:"dev"`
	PostgresHost     string `env:"POSTGRES_HOST,required"`
	PostgresPort     string `env:"POSTGRES_PORT,required"`
	PostgresDB       string `env:"POSTGRES_DB,required"`
	PostgresUser     string `env:"POSTGRES_USER,required"`
	PostgresPassword string `env:"POSTGRES_PASSWORD,required"`
}

var (
	cfg  *Config
	once sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		cfg = &Config{}
		err := env.Parse(cfg)

		if err != nil {
			panic(err)
		}
	})

	return cfg
}
