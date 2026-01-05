package config

import (
	"flag"
	"log"
	"os"
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
	FrontendURL      string `env:"FRONTEND_URL" envDefault:"localhost"`
	REST_APIPort     int    `env:"REST_API_PORT" envDefault:"8080"`
	LivePort         int    `env:"LIVE_PORT" envDefault:"8080"`
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
	SmtpHost         string `env:"SMTP_HOST" envDefault:"smtp.gmail.com"`
	SmtpPort         string `env:"SMTP_PORT" envDefault:"587"`
	SmtpUser         string `env:"SMTP_USER"`
	SmtpPass         string `env:"SMTP_PASS"`
	EmailFrom        string `env:"EMAIL_FROM" envDefault:"alerts@alerting.platform"`
}

var (
	cfg  *Config
	once sync.Once
)

func IsTesting() bool {
	return flag.Lookup("test.v") != nil
}

func SetTestEnv() {
	os.Setenv("SECRET", "secret")
	os.Setenv("POSTGRES_HOST", "localhost")
	os.Setenv("POSTGRES_PORT", "5432")
	os.Setenv("POSTGRES_DB", "testdb")
	os.Setenv("POSTGRES_USER", "testuser")
	os.Setenv("POSTGRES_PASSWORD", "testpass")
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")
	os.Setenv("REDIS_PASSWORD", "")
	os.Setenv("PROJECT_ID", "test-project")
}

func GetConfig() *Config {
	once.Do(func() {
		_ = godotenv.Load()

		if IsTesting() {
			SetTestEnv()
		}

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
