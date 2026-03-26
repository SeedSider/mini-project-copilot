package handlers

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

	"github.com/bankease/user-profile-service/internal/models"
)

// UploadHandler handles image upload to Azure Blob Storage.
type UploadHandler struct {
	AzureSASURL    string // account-level SAS URL, e.g. https://account.blob.core.windows.net/?sv=...&sig=...
	AzureContainer string // container name, e.g. "images"
}

var allowedMIME = map[string]string{
	"image/jpeg": "jpg",
	"image/png":  "png",
	"image/gif":  "gif",
	"image/webp": "webp",
}

// detectMIME returns the MIME type and file extension for the given content.
// SVG is handled separately because http.DetectContentType returns "text/xml"
// for XML-based formats rather than "image/svg+xml".
func detectMIME(data []byte) (mime, ext string, ok bool) {
	sniff := data
	if len(sniff) > 512 {
		sniff = sniff[:512]
	}
	detected := http.DetectContentType(sniff)

	if ext, found := allowedMIME[detected]; found {
		return detected, ext, true
	}

	// SVG detection: DetectContentType returns "text/xml" for SVG files.
	// Check the content for an <svg element to confirm.
	if detected == "text/xml; charset=utf-8" || detected == "text/plain; charset=utf-8" {
		sniffStr := strings.ToLower(strings.TrimSpace(string(sniff)))
		if strings.Contains(sniffStr, "<svg") {
			return "image/svg+xml", "svg", true
		}
	}

	return detected, "", false
}

// UploadImage handles POST /api/upload/image
// @Summary      Upload an image
// @Description  Upload an image file to Azure Blob Storage. Returns the URL of the uploaded image.
// @Tags         Upload
// @Accept       multipart/form-data
// @Produce      json
// @Param        image  formData  file  true  "Image file (jpeg, png, gif, webp — max 5MB)"
// @Success      200    {object}  models.UploadResponse
// @Failure      400    {object}  models.StandardResponse
// @Failure      413    {object}  models.StandardResponse
// @Failure      500    {object}  models.StandardResponse
// @Router       /api/upload/image [post]
func (h *UploadHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	if h.AzureSASURL == "" {
		writeError(w, http.StatusServiceUnavailable, "Image upload service is not configured")
		return
	}

	// Limit request body to 5MB before parsing multipart.
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

	// Read full content into memory (already limited to 5MB above).
	data, err := io.ReadAll(file)
	if err != nil {
		log.Printf("upload: failed to read file: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to read uploaded file")
		return
	}

	// Detect MIME type from content (not from filename) to prevent spoofing.
	mime, ext, ok := detectMIME(data)
	if !ok {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Unsupported image type '%s'. Allowed: jpeg, png, gif, webp, svg", mime))
		return
	}

	// Generate a random blob name to avoid collisions and path traversal.
	blobName, err := randomBlobName(ext)
	if err != nil {
		log.Printf("upload: failed to generate blob name: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Build the per-blob upload URL from the account SAS URL.
	// Account SAS URL: https://account.blob.core.windows.net/?sv=...&sig=...
	// Blob URL:        https://account.blob.core.windows.net/{container}/{blobName}?sv=...&sig=...
	blobURL, err := buildBlobURL(h.AzureSASURL, h.AzureContainer, blobName)
	if err != nil {
		log.Printf("upload: failed to build blob URL: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Upload via HTTP PUT (Azure REST API).
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

	writeJSON(w, http.StatusOK, models.UploadResponse{
		Code:        http.StatusOK,
		Description: "Image uploaded successfully",
		URL:         blobURL,
	})
}

// buildBlobURL constructs a blob-level SAS URL from an account-level SAS URL.
func buildBlobURL(sasURL, container, blobName string) (string, error) {
	u, err := url.Parse(sasURL)
	if err != nil {
		return "", fmt.Errorf("invalid AZURE_STORAGE_SAS_URL: %w", err)
	}
	// Replace path with /{container}/{blobName}, keep SAS query string.
	u.Path = fmt.Sprintf("/%s/%s", container, blobName)
	return u.String(), nil
}

// randomBlobName generates a cryptographically random hex filename with the given extension.
func randomBlobName(ext string) (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b) + "." + ext, nil
}
