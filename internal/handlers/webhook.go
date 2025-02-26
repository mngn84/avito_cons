package handlers

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/mngn84/avito-cons/internal/models/handlers_models"
	"github.com/mngn84/avito-cons/internal/services"
)

type WebhookHandler interface {
	HandleAvitoMsg(msg *handlers_models.FromAvitoMsg) (string, error) //error
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

func (h *webhookHandler) HandleAvitoMsg(msg *handlers_models.FromAvitoMsg) (string, error) {
	h.logger.Info("processing message", "msg", msg)

	itemInfo, err := h.avito.GetItemInfo(msg.UserId, msg.ChatId)
	if err != nil {
		h.logger.Error("failed to get item info", "error", err)
	}

	res, err := h.openai.GetResponse(msg.Content.Text, msg.ChatId, msg.UserId, msg.Created, itemInfo.Context.Value)

	if err != nil {
		return "", fmt.Errorf("failed to get response: %w", err) //fmt.Errorf("failed to get response: %w", err)
	}

	return res, nil //h.avito.SendMessage(msg.UserId, msg.ChatId, res)
}

func (h *webhookHandler) ServerHTTP(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 1048576)

	if r.Header.Get("Content-Type") != "application/json" {
		http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
		return
	}

	msg := handlers_models.FromAvitoMsg{}
	decoder := json.NewDecoder(r.Body)

	if err := decoder.Decode(&msg); err != nil {
		h.logger.Error("failed to decode request body", "error", err)
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if msg.AuthorId == 0 || msg.ChatId == "" {
		h.logger.Info("Invalid message received, skipping processing", "msg", msg)
		http.Error(w, "Invalid message data", http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	// Обрабатываем сообщение сразу (без горутины)
	resText, err := h.HandleAvitoMsg(&msg)
	if err != nil {
		h.logger.Error("failed to handle avito message", "error", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Отправляем ответ с текстом
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"response": resText})

	// go func() { // можно убрать после реализации очереди
	// 	if err := h.HandleAvitoMsg(&msg); err != nil {
	// 		h.logger.Error("failed to handle avito message", "error", err)
	// 	}
	// }()

}
