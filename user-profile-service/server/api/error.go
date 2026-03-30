package api

import (
	"encoding/json"
	"net/http"
)

// StandardResponse is the consistent response format for success/error.
type StandardResponse struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

// UploadResponse is returned after a successful image upload.
type UploadResponse struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
	URL         string `json:"url"`
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, StandardResponse{Code: status, Description: message})
}
