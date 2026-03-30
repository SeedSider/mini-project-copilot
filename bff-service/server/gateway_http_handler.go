package main

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type uploadResponse struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
	URL         string `json:"url,omitempty"`
}

// HandleUploadImage handles POST /api/upload/image (multipart/form-data).
func (s *gatewayServer) HandleUploadImage(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != "POST" {
		writeJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	azureSASURL := config.AzureSASURL
	azureContainer := config.AzureContainer

	if azureSASURL == "" {
		writeJSONError(w, http.StatusServiceUnavailable, "Azure SAS URL not configured")
		return
	}

	// Limit request body to 5MB
	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)

	if err := r.ParseMultipartForm(5 << 20); err != nil {
		writeJSONError(w, http.StatusRequestEntityTooLarge, "File too large (max 5MB)")
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Field 'image' is required")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to read file")
		return
	}

	// Detect MIME type
	mimeType := http.DetectContentType(data)
	allowedMIME := map[string]string{
		"image/jpeg":    ".jpg",
		"image/png":     ".png",
		"image/gif":     ".gif",
		"image/webp":    ".webp",
		"image/svg+xml": ".svg",
	}

	ext, ok := allowedMIME[mimeType]
	if !ok {
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("Unsupported MIME type: %s", mimeType))
		return
	}

	// Generate random filename
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to generate filename")
		return
	}
	filename := hex.EncodeToString(randomBytes) + ext

	// Build Azure URL
	parsedURL, err := url.Parse(azureSASURL)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Invalid Azure SAS URL configuration")
		return
	}

	blobPath := fmt.Sprintf("/%s/%s", azureContainer, filename)
	azureBlobURL := fmt.Sprintf("%s://%s%s?%s", parsedURL.Scheme, parsedURL.Host, blobPath, parsedURL.RawQuery)

	publicURL := fmt.Sprintf("%s://%s%s", parsedURL.Scheme, parsedURL.Host, blobPath)

	// Upload to Azure
	uploadReq, err := http.NewRequest("PUT", azureBlobURL, bytes.NewReader(data))
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to create upload request")
		return
	}
	uploadReq.Header.Set("x-ms-blob-type", "BlockBlob")
	uploadReq.Header.Set("Content-Type", mimeType)

	client := &http.Client{}
	resp, err := client.Do(uploadReq)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Failed to upload to Azure")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Azure upload failed with status %d", resp.StatusCode))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(uploadResponse{
		Code:        200,
		Description: "Image uploaded successfully",
		URL:         publicURL,
	})
}

func writeJSONError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error":   true,
		"code":    code,
		"message": message,
	})
}

// corsMiddleware applies CORS and security headers.
func corsMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Security headers
		w.Header().Set("Strict-Transport-Security", "max-age=31536000")
		w.Header().Set("Content-Security-Policy", "object-src 'none'; child-src 'none'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "no-referrer")

		// CORS headers
		w.Header().Set("Access-Control-Allow-Origin", strings.Join(config.CorsAllowedOrigins, ", "))
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.CorsAllowedMethods, ", "))
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.CorsAllowedHeaders, ", "))

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}

// grpcToHTTPCode maps gRPC status codes to HTTP status codes.
func grpcToHTTPCode(grpcCode int) int {
	switch grpcCode {
	case 0: // OK
		return 200
	case 3: // InvalidArgument
		return 422
	case 5: // NotFound
		return 404
	case 6: // AlreadyExists
		return 409
	case 7: // PermissionDenied
		return 403
	case 13: // Internal
		return 500
	case 14: // Unavailable
		return 503
	case 16: // Unauthenticated
		return 401
	default:
		return 500
	}
}
