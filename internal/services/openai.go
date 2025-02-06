package services

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/sashabaranov/go-openai"

	"github.com/mngn84/avito-cons/internal/config"
	"github.com/mngn84/avito-cons/internal/storage/pg"
)

type OpenAIService interface {
	GetResponse(text string, chatId string, userId int, created int) (string, error)
}

type openaiService struct {
	client *http.Client
	config *config.Config
	logger *slog.Logger
	db     *pg.SqliteClient/* PgClient */
}

func NewOpenAIService(config *config.Config, logger *slog.Logger, db *pg.SqliteClient/* PgClient */) OpenAIService {
	return &openaiService{
		client: &http.Client{},
		config: config,
		logger: logger,
		db:     db,
	}
}

func (s *openaiService) GetResponse(text string, chatId string, userId int, created int) (string, error) {
	client := openai.NewClient(s.config.OpenAI.ApiKey)
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: s.config.OpenAI.SystemPrompt,
		},
	}

	history, err := s.db.GetMessages(s.config.DB.HistoryLimit, chatId)
	if err != nil {
		s.logger.Error("Failed to get messages from database", "error", err)
	} else {
		for _, msg := range history {
			messages = append(messages, openai.ChatCompletionMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: text,
	})

	res, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       s.config.OpenAI.Model,
			Messages:    messages,
			Temperature: s.config.OpenAI.Temperature,
		},
	)

	if err != nil {
		s.logger.Error("Ошибка запроса к OpenAI", "error", err)
	}

	if len(res.Choices) == 0 {
		return "", fmt.Errorf("OpenAI вернул пустой ответ")
	}

	userMsg := pg.DbRow{
		ChatId:    chatId,
		UserId:    userId,
		Content:   text,
		Role:      "user",
		CreatedAt: created,
	}

	gptMsg := pg.DbRow{
		ChatId:    chatId,
		UserId:    userId,
		Content:   res.Choices[0].Message.Content,//"test content", //
		Role:      "assistant",
		CreatedAt: created,
	}

	s.logger.Info("userMsg", userMsg.Content, "gptMsg", gptMsg.Content, "error")

	err = s.db.SaveMsgPair(userMsg, gptMsg)
	if err != nil {
		s.logger.Error("Failed to save message to database", "error", err)
	}


	return res.Choices[0].Message.Content, nil
}
