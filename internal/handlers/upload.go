package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/mngn84/avito-cons/internal/services"
)

func UploadFileHandler(upload *services.UploadService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fileId, err := upload.UploadFile(r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"fileId": fileId})
	}
}