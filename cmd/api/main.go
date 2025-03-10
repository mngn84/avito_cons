package main

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"

	_ "github.com/golang-migrate/migrate/source/file"
	"github.com/mngn84/avito-cons/internal/config"
	"github.com/mngn84/avito-cons/internal/handlers"
	"github.com/mngn84/avito-cons/internal/services"
	"github.com/mngn84/avito-cons/internal/storage/pg"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file")
	}

	fmt.Printf("WEBHOOK_HOST=%s\nOPENAI_MODEL=%s\nOPENAI_URL=%s\n", os.Getenv("WEBHOOK_HOST"), os.Getenv("OPENAI_MODEL"), os.Getenv("OPENAI_URL"))

	cfg, err := config.New()
	if err != nil {
		log.Fatal("Config error: ", err)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	logger.Info("Starting server on ",
		"host", cfg.Webhook.Host,
		"port", cfg.Webhook.Port,
	)

	r := chi.NewRouter()

	db, err := pg.NewPgClient(cfg, logger)
	if err != nil {
		log.Fatal("DB error: ", err)
	}

	avito := services.NewAvitoService(cfg, logger)
	openai := services.NewOpenAIService(cfg, logger, db)
	upload := services.NewUploadService(openai, logger)
	h := handlers.NewWebhookHandler(avito, openai, logger)

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/webhook", h.ServerHTTP)
	r.Get("/health", handlers.HealthCheckHandler())
	r.Post("/upload", handlers.UploadFileHandler(upload))

	server := &http.Server{
		Addr:    ":" + cfg.Webhook.Port,
		Handler: r,
	}

	e := server.ListenAndServe()

	if e != nil {
		log.Fatal("Server error: ", e)
	}
}