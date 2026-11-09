package config

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

func ConnectDB() (*sql.DB, error) {

	dbUser := os.Getenv("ORDER_DB_USER")
	dbPass := os.Getenv("ORDER_DB_PASSWORD")
	dbHost := os.Getenv("ORDER_DB_HOST")
	dbPort := os.Getenv("ORDER_DB_PORT")
	dbName := os.Getenv("ORDER_DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&loc=Local", dbUser, dbPass, dbHost, dbPort, dbName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	err = db.Ping()

	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(50)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(3 * time.Minute)

	return db, nil
}
