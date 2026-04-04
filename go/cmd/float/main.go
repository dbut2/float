package main

import (
	"context"
	"database/sql"
	"log"
	"os"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"google.golang.org/api/option"

	floatDB "dbut.dev/float/db"
	"dbut.dev/float/go/api"
	"dbut.dev/float/go/database"
	"dbut.dev/float/go/frankfurter"
	"dbut.dev/float/go/middleware"
	"dbut.dev/float/go/seed"
	"dbut.dev/float/go/service"
	"dbut.dev/float/go/static"
	"dbut.dev/float/go/webhook"
)

func main() {
	r := gin.Default()
	apiHandler, auth := setup(r)

	apis := r.Group("/api")
	apis.Use(auth)
	apiHandler.Register(apis)

	static.Register(r)

	log.Println("starting server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}

func setup(r *gin.Engine) (*api.API, gin.HandlerFunc) {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_DSN"))
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	runMigrations(db)
	queries := database.New(db)

	fx := frankfurter.NewFXClient()

	var classifier *service.ClassifierService
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		model := os.Getenv("OPENAI_MODEL")
		if model == "" {
			model = "gpt-5.4-mini"
		}
		classifier = service.NewClassifierService(queries, service.NewOpenAIClient(apiKey, model))
	} else {
		log.Println("OPENAI_API_KEY not set, classification disabled")
	}

	push := service.NewPushService(queries, newFCMClient())

	if os.Getenv("DEMO_MODE") == "true" {
		user, buckets, txs, transfers, trickles := seed.DemoScenario()
		userID, err := service.LoadData(context.Background(), queries, user, buckets, txs, transfers, trickles)
		if err != nil {
			log.Fatalf("failed to seed demo data: %v", err)
		}
		return api.New(queries, fx, classifier, push), middleware.DemoAuth(userID)
	}

	webhook.New(queries, classifier, push).Register(r.Group("/webhook"))

	return api.New(queries, fx, classifier, push), middleware.Middleware(queries, os.Getenv("BASE_URL"))
}

func newFCMClient() *messaging.Client {
	creds := os.Getenv("FIREBASE_CREDENTIALS")
	if creds == "" {
		log.Println("FIREBASE_CREDENTIALS not set, push notifications disabled")
		return nil
	}
	app, err := firebase.NewApp(context.Background(), nil, option.WithAuthCredentialsJSON(option.ServiceAccount, []byte(creds)))
	if err != nil {
		log.Printf("firebase: failed to init app: %v", err)
		return nil
	}
	client, err := app.Messaging(context.Background())
	if err != nil {
		log.Printf("firebase: failed to init messaging: %v", err)
		return nil
	}
	return client
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
