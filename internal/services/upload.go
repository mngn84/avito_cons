package services

import (
	"fmt"
	"log/slog"
	"net/http"
	"path"
)

type UploadService struct {
	openai OpenAIService
	logger *slog.Logger
}

func NewUploadService(openai OpenAIService, logger *slog.Logger) *UploadService {
	return &UploadService{
		openai: openai,
		logger: logger,
	}
}

func (s *UploadService) UploadFile(r *http.Request) (string, error) {
	s.logger.Info("UploadFile")

	file, header, err := r.FormFile("file")
	if err != nil {
		return "", fmt.Errorf("failed to get file from request: %w", err)
	}

	profileName := r.FormValue("profile_name")
	if profileName == "" {
		return "", fmt.Errorf("profile name is required")
	}

	fileType := "instr"
	if path.Ext(header.Filename) == ".json" {
		fileType = "assrt"
	}
	defer file.Close()
	
	s.logger.Info("UploadFile", "fileType", fileType, "profileName", profileName, "fileName", header.Filename)

	return s.openai.UploadFileToVectorStore(file, header.Filename, profileName, fileType)
}
