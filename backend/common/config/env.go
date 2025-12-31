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
	Env              string `env:"ENV" envDefault:"dev"`
	Version          string `env:"VERSION" envDefault:"v0.1.0"`
	BuildTime        string `env:"BUILD_TIME" envDefault:"unknown"`
	Secret           string `env:"SECRET,required"`
	APIHost          string `env:"API_HOST" envDefault:"localhost"`
	REST_APIPort     int    `env:"REST_API_PORT" envDefault:"8080"`
	RPCPort          int    `env:"RPC_PORT" envDefault:"9090"`
	PostgresHost     string `env:"POSTGRES_HOST,required"`
	PostgresPort     string `env:"POSTGRES_PORT,required"`
	PostgresDB       string `env:"POSTGRES_DB,required"`
	PostgresUser     string `env:"POSTGRES_USER,required"`
	PostgresPassword string `env:"POSTGRES_PASSWORD,required"`
	RedisHost        string `env:"REDIS_HOST,required"`
	RedisPort        string `env:"REDIS_PORT,required"`
	RedisDB          int    `env:"REDIS_DB" envDefault:"0"`
	RedisPassword    string `env:"REDIS_PASSWORD,required"`
	RedisPrefix      string `env:"REDIS_PREFIX,required"`
	ProjectID        string `env:"PROJECT_ID,required"`
	FirestoreDB      string `env:"FIRESTORE_DB" envDefault:"logger-db"`
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

func Intro(name string) {
	config := GetConfig()
	log.Printf("Starting Alerting Platform %s - Version: %s, Build Time: %s, Environment: %s", name, config.Version, config.BuildTime, config.Env)
}
