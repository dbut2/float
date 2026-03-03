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
	apiHandler, auth := setup()

	r := gin.Default()
	apis := r.Group("/api")
	apis.Use(auth)
	apiHandler.Register(apis)

	log.Println("starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func setup() (*api.API, gin.HandlerFunc) {
	if os.Getenv("DEMO_MODE") == "true" {
		apiHandler, userID := api.NewDemo()
		return apiHandler, middleware.DemoAuth(userID)
	}

	db, err := sql.Open("postgres", os.Getenv("DATABASE_DSN"))
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	runMigrations(db)
	queries := database.New(db)
	return api.New(queries), middleware.Auth(queries)
}

func runMigrations(db *sql.DB) {
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
}
