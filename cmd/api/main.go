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

	"github.com/mngn84/avito-cons/internal/config"
	"github.com/mngn84/avito-cons/internal/handlers"
	"github.com/mngn84/avito-cons/internal/services"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Error loading .env file")
	}

	fmt.Printf("WEBHOOK_HOST=%s\n", os.Getenv("WEBHOOK_HOST"))

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

	aserv := services.NewAvitoService(cfg, logger)
	oserv := services.NewOpenAIService(cfg, logger)
	h := handlers.NewWebhookHandler(aserv, oserv, logger)

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/webhook", h.ServerHTTP)
	r.Get("/health", handlers.HealthCheckHandler())

	server := &http.Server{
		Addr:    ":" + cfg.Webhook.Port,
		Handler: r,
	}

	e := server.ListenAndServe()

	if e != nil {
		log.Fatal("Server error: ", e)
	}
}
