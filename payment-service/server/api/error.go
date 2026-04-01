package api

import (
	"encoding/json"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type standardResponse struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

type errorResponse struct {
	Error   bool   `json:"error"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, standardResponse{Code: statusCode, Description: message})
}

func writeAuthError(w http.ResponseWriter, statusCode int, message string) {
	writeJSON(w, statusCode, errorResponse{Error: true, Code: statusCode, Message: message})
}

func (s *Server) serverError() error {
	return status.New(codes.Internal, "Internal Error").Err()
}

func (s *Server) unauthorizedError() error {
	return status.New(codes.Unauthenticated, "Unauthorized").Err()
}

type prepaidErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

func writePrepaidError(w http.ResponseWriter, statusCode int, errorCode, message string) {
	writeJSON(w, statusCode, prepaidErrorResponse{Error: errorCode, Message: message})
}
