package handlers

import (
	"encoding/json"
	//"fmt"
	"log/slog"
	"net/http"

	"github.com/mngn84/avito-cons/internal/models/handlers_models"
	"github.com/mngn84/avito-cons/internal/services"
)

type WebhookHandler interface {
	HandleAvitoMsg(msg *handlers_models.FromAvitoMsg) error
	ServerHTTP(w http.ResponseWriter, r *http.Request)
}

type webhookHandler struct {
	avito  services.AvitoService
	openai services.OpenAIService
	logger *slog.Logger
}

func NewWebhookHandler(avito services.AvitoService, openai services.OpenAIService, logger *slog.Logger) WebhookHandler {
	return &webhookHandler{
		avito:  avito,
		openai: openai,
		logger: logger,
	}
}

func (h *webhookHandler) HandleAvitoMsg(msg *handlers_models.FromAvitoMsg) error {
	h.logger.Info("processing message", "msg", msg)

	// res, err := h.openai.GetResponse(msg.Content.Text)
	// if err != nil {
	// 	return fmt.Errorf("failed to get response: %w", err)
	// }

	return nil // h.avito.SendMessage(msg.UserId, msg.ChatId, res)
}

func (h *webhookHandler) ServerHTTP(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	var msg handlers_models.FromAvitoMsg
	decoder := json.NewDecoder(r.Body)
	// decoder.DisallowUnknownFields() // проверка на наличие неизвестных полей

	if err := decoder.Decode(&msg); err != nil {
		h.logger.Error("failed to decode request body", "error", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)

	go func() { // можно убрать после реализации очереди
		if err := h.HandleAvitoMsg(&msg); err != nil {
			h.logger.Error("failed to handle avito message", "error", err)
		}
	}()

}
