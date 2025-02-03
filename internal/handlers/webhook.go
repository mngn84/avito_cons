package handlers

import (
	"encoding/json"
	"fmt"
	//"fmt"
	"log/slog"
	"net/http"

	"github.com/mngn84/avito-cons/internal/models/handlers_models"
	"github.com/mngn84/avito-cons/internal/services"
)

type WebhookHandler interface {
	HandleAvitoMsg(w http.ResponseWriter, msg *handlers_models.FromAvitoMsg) error
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

func (h *webhookHandler) HandleAvitoMsg(w http.ResponseWriter, msg *handlers_models.FromAvitoMsg) error {
	h.logger.Info("processing message", "msg", msg)

	resText, err := h.openai.GetResponse(msg.Content.Text, msg.ChatId, msg.UserId, msg.Created)
	if err != nil {
		return fmt.Errorf("failed to get response: %w", err)
	}
	// Кодируем результат в JSON и отправляем его в ответ
	w.Header().Set("Content-Type", "application/json")                       //test
	w.WriteHeader(http.StatusOK)                                             //test
	return json.NewEncoder(w).Encode(map[string]string{"response": resText}) //test
	// return nil  //h.avito.SendMessage(msg.UserId, msg.ChatId, resText)
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

	//w.WriteHeader(http.StatusOK)

	go func() { // можно убрать после реализации очереди
		if err := h.HandleAvitoMsg(w, &msg); err != nil {
			h.logger.Error("failed to handle avito message", "error", err)
		}
	}()

}
