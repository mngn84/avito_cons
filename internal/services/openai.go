package services

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/sashabaranov/go-openai"

	"github.com/mngn84/avito-cons/internal/config"
	"github.com/mngn84/avito-cons/internal/models/avito_models"
	"github.com/mngn84/avito-cons/internal/storage/pg"
)

type OpenAIService interface {
	GetResponse(text string, chatId string, userId int, created int, itemInfo avito_models.Value) (string, error)
	UploadFileToVectorStore(file io.Reader, fileName, profileName, fileType string) (string, error)
}

type openaiService struct {
	client *http.Client
	config *config.Config
	logger *slog.Logger
	db     *pg.PgClient
	openai *openai.Client
	ctx    context.Context
}

func NewOpenAIService(config *config.Config, logger *slog.Logger, db *pg.PgClient) OpenAIService {
	ctx := context.Background()
	return &openaiService{
		client: &http.Client{},
		config: config,
		logger: logger,
		db:     db,
		openai: openai.NewClient(config.OpenAI.ApiKey),
		ctx:    ctx,
	}
}

func (s *openaiService) GetResponse(text string, chatId string, userId int, created int, itemInfo avito_models.Value) (string, error) {
	asstId, err := s.getAssistantId(userId)
	if err != nil {
		return "", err
	}

	threadId, isNew, err := s.getOrCreateThread(chatId, asstId)
	if err != nil {
		return "", err
	}

	err = s.sendMessageToThread(threadId, text, itemInfo, isNew)
	if err != nil {
		return "", err
	}

	runId, err := s.runAssistant(threadId, asstId)
	if err != nil {
		return "", err
	}

	res, err := s.waitForResponse(threadId, runId)
	if err != nil {
		return "", err
	}

	return res, nil
}

func (s *openaiService) getAssistantId(userId int) (string, error) {
	asstId, err := s.db.GetAssistantId(userId)
	if err != nil || asstId == "" {
		return "", fmt.Errorf("failed to get assistant id: %w", err)
	}

	return asstId, nil
}

func (s *openaiService) getOrCreateThread(chatId, asstId string) (string, bool, error) {
	threadId, err := s.db.GetThreadId(chatId)
	if err == nil && threadId != "" {
		return threadId, false, nil
	}

	thread, err := s.openai.CreateThread(s.ctx, openai.ThreadRequest{})
	if err != nil {
		s.logger.Error("failed to create thread", "error", err)
		return "", false, err
	}

	_ = s.db.SaveThreadId(chatId, thread.ID, asstId)
	return thread.ID, true, nil
}

func (s *openaiService) sendMessageToThread(threadId, text string, itemInfo avito_models.Value, isNew bool) error {
	if isNew {
		text = fmt.Sprintf("Сообщение по объявлению %s %s: %s", itemInfo.Title, itemInfo.PriceString, text)
	}
	_, err := s.openai.CreateMessage(s.ctx, threadId, openai.MessageRequest{
		Role:    "user",
		Content: text,
	})

	if err != nil {
		s.logger.Error("failed to create message", "error", err)
		return err
	}

	return nil
}

func (s *openaiService) runAssistant(threadId string, asstId string) (string, error) {
	run, err := s.openai.CreateRun(s.ctx, threadId, openai.RunRequest{
		AssistantID: asstId,
	})
	if err != nil {
		s.logger.Error("failed to create run", "error", err)
		return "", err
	}
	return run.ID, nil
}

func (s *openaiService) waitForResponse(threadId string, runId string) (string, error) {
	for {
		res, err := s.openai.RetrieveRun(s.ctx, threadId, runId)
		if err != nil {
			return "", fmt.Errorf("failed to get run status: %w", err)
		}

		switch res.Status {
		case openai.RunStatusCompleted:
			limit := 1
			order := "desc"

			msgs, err := s.openai.ListMessage(s.ctx, threadId, &limit, &order, nil, nil, nil)
			if err != nil {
				return "", fmt.Errorf("failed to get message: %w", err)
			}

			if len(msgs.Messages) == 0 {
				return "", fmt.Errorf("no messages found")
			}

			return msgs.Messages[0].Content[0].Text.Value, nil
		case openai.RunStatusFailed:
			return "", fmt.Errorf("failed to run assistant")
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func (s *openaiService) UploadFileToVectorStore(file io.Reader, fileName, profileName, fileType string) (string, error) {
	s.logger.Info("Uploading file to vector store")

	userId, err := s.db.GetUserId(profileName)
	if err != nil {
		return "", fmt.Errorf("failed to get user id: %w", err)
	}

	asstId, err := s.getAssistantId(userId)
	if err != nil || asstId == "" {
		asstId, err = s.createAssistant(userId, profileName)
		if err != nil {
			return "", fmt.Errorf("failed to create assistant: %w", err)
		}
	}

	storeId, err := s.db.GetStoreId(asstId)
	if err != nil || storeId == "" {
		s.logger.Info("Vector store not found")
		storeId, err = s.createVectorStore(asstId, profileName)
		if err != nil {
			return "", fmt.Errorf("failed to create vector store: %w", err)
		}
	}
	s.logger.Info("Vector store", "store_id", storeId)

	oldFileId, err := s.db.GetOldFileId(storeId, fileName, fileType)

	if err == nil && oldFileId != "" {
		err = s.deleteOldFile(storeId, oldFileId)
		if err != nil {
			return "", fmt.Errorf("failed to delete old file: %w", err)
		}
	}

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	fileResp, err := s.openai.CreateFileBytes(s.ctx, openai.FileBytesRequest{
		Name:    fileName,
		Bytes:   fileBytes,
		Purpose: "assistants",
	})
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	fileId := fileResp.ID
	s.logger.Info("File uploaded to openai", "file_id", fileId)

	err = s.addFileToStore(fileId, fileName, fileType, storeId)
	if err != nil {
		return "", err
	}

	return fileId, nil
}

func (s *openaiService) createAssistant(userId int, profileName string) (string, error) {
	s.logger.Info("Creating assistant")

	asstId, err := s.getAssistantId(userId)
	if err == nil && asstId != "" {
		return "", err
	}

	asstName := fmt.Sprintf("%s-asst", profileName)
	asst, err := s.openai.CreateAssistant(s.ctx, openai.AssistantRequest{
		Model:        s.config.OpenAI.Model,
		Name:         &asstName,
		Instructions: &s.config.OpenAI.SystemPrompt,
		//tools: ????
	})

	if err != nil {
		s.logger.Error("failed to create assistant", "error", err)
		return "", err
	}

	_ = s.db.SaveAssistant(asst.ID, *asst.Name, userId)
	return asst.ID, nil
}

func (s *openaiService) createVectorStore(asstId, profileName string) (string, error) {
	storeName := fmt.Sprintf("vector-store_%s", profileName)
	s.logger.Info("Creating vector store", "store_name", storeName)

	store, err := s.openai.CreateVectorStore(s.ctx, openai.VectorStoreRequest{
		Name: storeName,
	})
	if err != nil {
		return "", fmt.Errorf("failed to create vector store: %w", err)
	}

	err = s.db.SaveStoreRecord(store.ID, store.Name, asstId)
	if err != nil {
		return "", fmt.Errorf("failed to save vector store id: %w", err)
	}

	return store.ID, nil
}

func (s *openaiService) addFileToStore(fileId, fileName, fileType, storeId string) error {
	s.logger.Info("Adding file to vector store", "file_id", fileId)

	file, err := s.openai.CreateVectorStoreFile(s.ctx, storeId, openai.VectorStoreFileRequest{
		FileID: fileId,
	})
	if err != nil {
		return fmt.Errorf("failed to add file to vector store: %w", err)
	}

	err = s.db.SaveFileRecord(file.ID, fileName, fileType, file.VectorStoreID)
	if err != nil {
		return fmt.Errorf("failed to save file id: %w", err)
	}

	return nil
}

func (s *openaiService) deleteOldFile(storeId, fileId string) error {
	err := s.openai.DeleteVectorStoreFile(s.ctx, storeId, fileId)
	if err != nil {
		return fmt.Errorf("failed to delete vector store files: %w", err)
	}

	err = s.openai.DeleteFile(s.ctx, fileId)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	err = s.db.DeleteOldFile(storeId, fileId)
	if err != nil {
		return fmt.Errorf("failed to delete file record: %w", err)
	}

	return nil
}

//удаление файлов из векторного хранилища
/* func (s *openaiService) listVectorStores() (openai.VectorStoresList, error) {
	stores, err := s.openai.ListVectorStores(s.ctx, nil, l)
	if err != nil {
		return nil, fmt.Errorf("failed to list vector stores: %w", err)
	}
	return stores, nil
}
*/
