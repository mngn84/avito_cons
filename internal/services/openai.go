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
	db     *pg.SqliteClient /* PgClient */
}

func NewOpenAIService(config *config.Config, logger *slog.Logger, db *pg.SqliteClient /* PgClient */) OpenAIService {
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

		rows, err := s.db.DB().Query("SELECT * FROM messages WHERE chat_id = $1", chatId)
		if err != nil {
			s.logger.Error("Ошибка запроса к БД", "error", err)
		} else {
			defer rows.Close()
			for rows.Next() {
				var msg pg.DbRow
				if err := rows.Scan(&msg.UserId, &msg.ChatId, &msg.Content, &msg.Role, &msg.CreatedAt); err != nil {
					s.logger.Error("Ошибка сканирования строки", "error", err)
				}
			}
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
		Content:   res.Choices[0].Message.Content,/*  "test content",  */
		Role:      "assistant",
		CreatedAt: created,
	}

	s.logger.Info("userMsg", userMsg.Content, "gptMsg", gptMsg.Content, "error")

	err = s.db.SaveMsgPair(userMsg, gptMsg)
	if err != nil {
		s.logger.Error("Failed to save message to database", "error", err)
	}
	
	return res.Choices[0].Message.Content/* "test content" */, nil
}

/* func (s *openaiService) GetResponse(text string, chatId string, userId int, created int) (string, error) {
	// Получаем ответ от OpenAI с помощью прямого HTTP запроса
	url := s.config.OpenAI.ApiUrl + "/v1/chat/completions"
	s.logger.Info("Запрос к ", "url ", url)
	apiKey := s.config.OpenAI.ApiKey // API ключ из конфигурации
	model := s.config.OpenAI.Model   // Модель, например "gpt-3.5-turbo" или "gpt-4"
	// Отправляем запрос к OpenAI через прямое соединение
	response, err := sendOpenAIRequest(url, apiKey, model, text)
	if err != nil {
		s.logger.Error("Ошибка запроса к OpenAI", "error", err)
		return "", err
	}

	// Здесь можно добавить логику для сохранения сообщений в базе данных или другие шаги
	return response, nil
}
func sendOpenAIRequest(url, apiKey, model, content string) (string, error) {

	// Подготовка данных запроса
	requestData := map[string]interface{}{
		"model": model,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": content,
			},
		},
	}

	// Кодируем в JSON
	jsonData, err := json.Marshal(requestData)
	if err != nil {
		return "", err
	}

	// Отправляем запрос
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}

	// Добавляем заголовки
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer " + apiKey)

	// Отправляем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Читаем ответ
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	// Пытаемся разобрать ответ
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error, status code: %d, status: %s, message: %s", resp.StatusCode, resp.Status, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	// Получаем текст ответа из OpenAI
	choices, ok := result["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return "", fmt.Errorf("no choices found in OpenAI response")
	}

	// Извлекаем ответ из первого выбора
	responseText, ok := choices[0].(map[string]interface{})["message"].(map[string]interface{})["content"].(string)
	if !ok {
		return "", fmt.Errorf("invalid response format")
	}

	return responseText, nil
}
*/
