package services

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/mngn84/avito-cons/internal/config"
	"github.com/sashabaranov/go-openai"
)

type OpenAIService interface {
	GetResponse(text string) (string, error)
}

type openaiService struct {
	client *http.Client
	config *config.Config
	logger *slog.Logger
}

func NewOpenAIService(config *config.Config, logger *slog.Logger) OpenAIService {
	return &openaiService{
		client: &http.Client{},
		config: config,
		logger: logger,
	}
}

func (s *openaiService) GetResponse(text string) (string, error) {
	client := openai.NewClient(s.config.OpenAI.ApiKey)
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: s.config.OpenAI.SystemPrompt,
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: text,
		},
	}
	res, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: messages,
			Temperature: s.config.OpenAI.Temperature,
		},
	)

	if err != nil {
		return "", fmt.Errorf("OpenAI request failed: %w", err)
	}
	
	return res.Choices[0].Message.Content, nil
}
