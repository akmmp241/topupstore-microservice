package shared

import (
	"database/sql"
	_ "github.com/go-sql-driver/mysql"
	"log/slog"
	"os"
	"time"
)

func GetConnection() *sql.DB {
	dbName := os.Getenv("DB_NAME")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")

	db, err := sql.Open("mysql", dbUser+":"+dbPassword+"@tcp("+dbHost+":"+dbPort+")/"+dbName+"?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}

	db.SetMaxIdleConns(10)
	db.SetMaxOpenConns(100)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(60 * time.Minute)

	return db
}

func CommitOrRollback(tx *sql.Tx, err error) error {
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return rollbackErr
		}
		return err
	}
	if commitErr := tx.Commit(); commitErr != nil {
		return commitErr
	}
	return nil
}
