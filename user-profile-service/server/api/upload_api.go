package api

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

var allowedMIME = map[string]string{
	"image/jpeg": "jpg",
	"image/png":  "png",
	"image/gif":  "gif",
	"image/webp": "webp",
}

// detectMIME returns the MIME type and file extension for the given content.
func detectMIME(data []byte) (mime, ext string, ok bool) {
	sniff := data
	if len(sniff) > 512 {
		sniff = sniff[:512]
	}
	detected := http.DetectContentType(sniff)

	if ext, found := allowedMIME[detected]; found {
		return detected, ext, true
	}

	if detected == "text/xml; charset=utf-8" || detected == "text/plain; charset=utf-8" {
		sniffStr := strings.ToLower(strings.TrimSpace(string(sniff)))
		if strings.Contains(sniffStr, "<svg") {
			return "image/svg+xml", "svg", true
		}
	}

	return detected, "", false
}

// HandleUploadImage handles POST /api/upload/image
func (s *Server) HandleUploadImage(w http.ResponseWriter, r *http.Request) {
	if s.azureSASURL == "" {
		writeError(w, http.StatusServiceUnavailable, "Image upload service is not configured")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)

	if err := r.ParseMultipartForm(5 << 20); err != nil {
		if err.Error() == "http: request body too large" {
			writeError(w, http.StatusRequestEntityTooLarge, "Image must be smaller than 5MB")
			return
		}
		writeError(w, http.StatusBadRequest, "Invalid multipart form")
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		writeError(w, http.StatusBadRequest, "Field 'image' is required")
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		log.Printf("upload: failed to read file: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to read uploaded file")
		return
	}

	mime, ext, ok := detectMIME(data)
	if !ok {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Unsupported image type '%s'. Allowed: jpeg, png, gif, webp, svg", mime))
		return
	}

	blobName, err := randomBlobName(ext)
	if err != nil {
		log.Printf("upload: failed to generate blob name: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	blobURL, err := buildBlobURL(s.azureSASURL, s.azureContainer, blobName)
	if err != nil {
		log.Printf("upload: failed to build blob URL: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	req, err := http.NewRequestWithContext(r.Context(), http.MethodPut, blobURL, bytes.NewReader(data))
	if err != nil {
		log.Printf("upload: failed to create PUT request: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	req.Header.Set("x-ms-blob-type", "BlockBlob")
	req.Header.Set("Content-Type", mime)
	req.ContentLength = int64(len(data))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("upload: Azure PUT request failed: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to upload image to storage")
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		log.Printf("upload: Azure returned %d: %s", resp.StatusCode, string(body))
		writeError(w, http.StatusInternalServerError, "Storage service returned an error")
		return
	}

	writeJSON(w, http.StatusOK, UploadResponse{
		Code:        http.StatusOK,
		Description: "Image uploaded successfully",
		URL:         blobURL,
	})
}

func buildBlobURL(sasURL, container, blobName string) (string, error) {
	u, err := url.Parse(sasURL)
	if err != nil {
		return "", fmt.Errorf("invalid AZURE_STORAGE_SAS_URL: %w", err)
	}
	u.Path = fmt.Sprintf("/%s/%s", container, blobName)
	return u.String(), nil
}

func randomBlobName(ext string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b) + "." + ext, nil
}
