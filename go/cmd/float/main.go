package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"

	floatDB "dbut.dev/float/db"
	"dbut.dev/float/go/api"
	"dbut.dev/float/go/database"
	"dbut.dev/float/go/middleware"
)

func main() {
	databaseDSN := os.Getenv("DATABASE_DSN")

	db, err := sql.Open("postgres", databaseDSN)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer func() { _ = db.Close() }()
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	if _, err := db.Exec("CREATE SCHEMA IF NOT EXISTS float"); err != nil {
		log.Fatalf("failed to create schema: %v", err)
	}

	goose.SetBaseFS(floatDB.Migrations)
	goose.SetTableName("float.goose_db_version")
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("failed to set goose dialect: %v", err)
	}
	if err := goose.Up(db, "."); err != nil {
		log.Fatalf("failed to run database migrations: %v", err)
	}

	queries := database.New(db)
	apiHandler := api.New(queries)

	r := gin.Default()

	apis := r.Group("/api")
	apis.Use(middleware.Auth(queries))
	apiHandler.Register(apis)

	log.Println("starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
