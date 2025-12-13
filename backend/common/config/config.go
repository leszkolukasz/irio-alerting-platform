package config

import (
	"log"
	"sync"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

const DEV = "dev"
const PROD = "prod"

type Config struct {
	Env              string `env:"ENV" default:"dev"`
	Secret           string `env:"SECRET,required"`
	Port             int    `env:"PORT" default:"8080"`
	PostgresHost     string `env:"POSTGRES_HOST,required"`
	PostgresPort     string `env:"POSTGRES_PORT,required"`
	PostgresDB       string `env:"POSTGRES_DB,required"`
	PostgresUser     string `env:"POSTGRES_USER,required"`
	PostgresPassword string `env:"POSTGRES_PASSWORD,required"`
	RedisHost        string `env:"REDIS_HOST,required"`
	RedisPort        string `env:"REDIS_PORT,required"`
	RedisDB          int    `env:"REDIS_DB" default:"0"`
	RedisPassword    string `env:"REDIS_PASSWORD" required:""`
	RedisPrefix      string `env:"REDIS_PREFIX" required:""`
}

var (
	cfg  *Config
	once sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		_ = godotenv.Load()

		cfg = &Config{}
		err := env.Parse(cfg)

		if err != nil {
			log.Fatal("Failed to parse env vars: ", err)
		}
	})

	return cfg
}
