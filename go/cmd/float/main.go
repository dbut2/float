package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/pressly/goose/v3"

	floatDB "dbut.dev/float/db"
)

func main() {
	databaseDSN := os.Getenv("DATABASE_DSN")

	db, err := sql.Open("postgres", databaseDSN)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	goose.SetBaseFS(floatDB.Migrations)
	goose.SetTableName("bank.goose_db_version")
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("failed to set goose dialect: %v", err)
	}
	if err := goose.Up(db, "."); err != nil {
		log.Fatalf("failed to run database migrations: %v", err)
	}
}
