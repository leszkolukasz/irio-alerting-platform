package db

import (
	"alerting-platform/common/config"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func GetRawDBConnection() *sql.DB {
	var db *sql.DB
	var err error

	if config.GetConfig().Env == config.DEV {
		db, err = connectLocal()
	} else {
		db, err = connectLocal()
	}

	if err != nil {
		panic(fmt.Sprintf("Failed to connect to DB: %v", err))
	}

	return db
}

func connectLocal() (*sql.DB, error) {
	config := config.GetConfig()

	connectionString := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s",
		config.PostgresHost,
		config.PostgresUser,
		config.PostgresPassword,
		config.PostgresDB,
		config.PostgresPort,
	)

	return sql.Open("pgx", connectionString)
}

// func connectCloudSQL() (*sql.DB, error) {
// }
