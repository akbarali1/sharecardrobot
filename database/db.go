package database

import (
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var db *sql.DB

func buildDSN() string {
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "root"
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "root"
	}

	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "127.0.0.1"
	}

	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "3306"
	}

	name := os.Getenv("DB_DATABASE")
	if name == "" {
		name = "share_card_robot"
	}

	return fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Asia%%2FTashkent",
		user, password, host, port, name)
}

func InitConnection() error {
	dsn := buildDSN()

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		return err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	return db.Ping()
}

func CloseConnection() {
	if db != nil {
		_ = db.Close()
	}
}

func QueryRow(query string, params ...any) *sql.Row {
	return db.QueryRow(query, params...)
}

func Query(query string, params ...any) (*sql.Rows, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return db.Query(query, params...)
}

func Exec(query string, params ...any) (sql.Result, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}
	return db.Exec(query, params...)
}
