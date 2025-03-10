package services

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	stdhttp "net/http"
	"time"

	"github.com/mngn84/avito-cons/internal/config"
	"github.com/mngn84/avito-cons/internal/http"
	"github.com/mngn84/avito-cons/internal/models/avito_models"
)

type AvitoService interface {
	SendMessage(userId int, chatId string, text string) error
	ReadChat(userId int, chatId string) error
	GetItemInfo(userId int, chatId string) (avito_models.GetChatInfoResponse, error)
}

type avitoService struct {
	client *http.Client
	config *config.Config
	logger *slog.Logger
}

func NewAvitoService(config *config.Config, logger *slog.Logger) AvitoService {
	httpClient := &stdhttp.Client{
		Timeout: 10 * time.Second,
	}

	retryConfig := http.RetryConfig{
		MaxRetries: 3,
		BaseDelay:  1 * time.Second,
		MaxDelay:   5 * time.Second,
	}

	customClient := http.NewClient(httpClient, logger, retryConfig)

	return &avitoService{
		client: customClient,
		config: config,
		logger: logger,
	}
}


func (s *avitoService) SendMessage(userId int, chatId string, text string) error {
	ctx := context.Background()

	msg := avito_models.ToAvitoMsg{
		Message: avito_models.Msg{
			Text: text,
		},
		Type: "text",
	}

	jsonData, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal json: %w", err)
	}

	url := fmt.Sprintf("%s/messenger/v1/accounts/%d/chats/%s/messages", s.config.Avito.ApiUrl, userId, chatId)

	req, err := stdhttp.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.config.Avito.Token))

	body, err := s.client.Do(ctx, req)
	if err != nil {
		if s.logger != nil {
            s.logger.Error("failed to send request", "error", err)
        }
		return fmt.Errorf("failed to send request: %w", err)
	}

	res := avito_models.SendMsgResponse{}
	if err := json.Unmarshal(body, &res); err != nil {
		if s.logger != nil {
            s.logger.Error("failed to unmarshal response", "error", err)
        }
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

func (s *avitoService) ReadChat(userId int, chatId string) error {
	ctx := context.Background()

	url := fmt.Sprintf("%s/messenger/v1/accounts/%d/chats/%s/read", s.config.Avito.ApiUrl, userId, chatId)

	req, err := stdhttp.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.config.Avito.Token))

	body, err := s.client.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	var res avito_models.ReadChatResponse
	if err := json.Unmarshal(body, &res); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}


func (s *avitoService) GetItemInfo(userId int, chatId string) (avito_models.GetChatInfoResponse, error) {
    ctx := context.Background()
    url := fmt.Sprintf("%s/messenger/v2/accounts/%d/chats/%s", s.config.Avito.ApiUrl, userId, chatId)

    req, err := stdhttp.NewRequestWithContext(ctx, "GET", url, stdhttp.NoBody)
    if err != nil {
        return avito_models.GetChatInfoResponse{}, fmt.Errorf("failed to create request: %w", err)
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.config.Avito.Token))

    // s.logger.Info("GetItemInfo", "method", req.Method, "url", req.URL.String(), "headers", req.Header)

    body, err := s.client.Do(ctx, req)
    if err != nil {
        // s.logger.Error("GetItemInfo", "failed to send request", err)
        return avito_models.GetChatInfoResponse{}, fmt.Errorf("failed to send request: %w", err)
    }

    if body == nil {
        // s.logger.Error("Ошибка: body == nil")
        return avito_models.GetChatInfoResponse{}, errors.New("response body is nil")
    }

    res := avito_models.GetChatInfoResponse{}
    if err := json.Unmarshal(body, &res); err != nil {
        return avito_models.GetChatInfoResponse{}, fmt.Errorf("failed to unmarshal response: %w", err)
    }

    return res, nil
}