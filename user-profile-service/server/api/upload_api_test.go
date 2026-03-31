package api

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ═══════════════════════════════════════════
// detectMIME
// ═══════════════════════════════════════════

func TestDetectMIME_JPEG(t *testing.T) {
	// JPEG magic bytes: FF D8 FF
	data := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10}
	mime, ext, ok := detectMIME(data)
	assert.True(t, ok)
	assert.Equal(t, "image/jpeg", mime)
	assert.Equal(t, "jpg", ext)
}

func TestDetectMIME_PNG(t *testing.T) {
	// PNG magic bytes: 89 50 4E 47
	data := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	mime, ext, ok := detectMIME(data)
	assert.True(t, ok)
	assert.Equal(t, "image/png", mime)
	assert.Equal(t, "png", ext)
}

func TestDetectMIME_SVG(t *testing.T) {
	data := []byte(`<svg xmlns="http://www.w3.org/2000/svg"><rect width="10" height="10"/></svg>`)
	mime, ext, ok := detectMIME(data)
	assert.True(t, ok)
	assert.Equal(t, "image/svg+xml", mime)
	assert.Equal(t, "svg", ext)
}

func TestDetectMIME_Unsupported(t *testing.T) {
	data := []byte("plain text content that is not an image")
	_, _, ok := detectMIME(data)
	assert.False(t, ok)
}

func TestDetectMIME_LargeData(t *testing.T) {
	// Data larger than 512 bytes — should only sniff first 512
	data := make([]byte, 1024)
	data[0] = 0xFF
	data[1] = 0xD8
	data[2] = 0xFF
	_, _, ok := detectMIME(data)
	// Partial JPEG header without proper content — just verify it doesn't panic
	_ = ok
}

// ═══════════════════════════════════════════
// buildBlobURL
// ═══════════════════════════════════════════

func TestBuildBlobURL_Success(t *testing.T) {
	url, err := buildBlobURL("https://storage.example.com?sv=token", "images", "abc123.jpg")
	require.NoError(t, err)
	assert.Contains(t, url, "/images/abc123.jpg")
	assert.Contains(t, url, "sv=token")
}

func TestBuildBlobURL_InvalidURL(t *testing.T) {
	_, err := buildBlobURL("://invalid-url", "images", "abc.jpg")
	assert.Error(t, err)
}

// ═══════════════════════════════════════════
// randomBlobName
// ═══════════════════════════════════════════

func TestRandomBlobName_Success(t *testing.T) {
	name, err := randomBlobName("jpg")
	require.NoError(t, err)
	assert.True(t, strings.HasSuffix(name, ".jpg"))
	assert.Greater(t, len(name), 4)
}

func TestRandomBlobName_DifferentExt(t *testing.T) {
	name, err := randomBlobName("png")
	require.NoError(t, err)
	assert.True(t, strings.HasSuffix(name, ".png"))
}

func TestRandomBlobName_Unique(t *testing.T) {
	name1, _ := randomBlobName("jpg")
	name2, _ := randomBlobName("jpg")
	assert.NotEqual(t, name1, name2)
}

// ═══════════════════════════════════════════
// HandleUploadImage — early-return paths
// ═══════════════════════════════════════════

func TestHandleUploadImage_NotConfigured(t *testing.T) {
	srv, _ := newTestServer(t) // azureSASURL = ""
	r := httptest.NewRequest(http.MethodPost, "/api/upload/image", nil)
	w := httptest.NewRecorder()

	srv.HandleUploadImage(w, r)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
}

func TestHandleUploadImage_InvalidMultipart(t *testing.T) {
	// Server with azureSASURL set so it passes the config check
	srv := &Server{azureSASURL: "https://storage.example.com", azureContainer: "images"}
	r := httptest.NewRequest(http.MethodPost, "/api/upload/image", strings.NewReader("not-multipart"))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	srv.HandleUploadImage(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleUploadImage_MissingImageField(t *testing.T) {
	srv := &Server{azureSASURL: "https://storage.example.com", azureContainer: "images"}

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	mw.WriteField("other_field", "value")
	mw.Close()

	r := httptest.NewRequest(http.MethodPost, "/api/upload/image", &buf)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()

	srv.HandleUploadImage(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandleUploadImage_UnsupportedMIME(t *testing.T) {
	srv := &Server{azureSASURL: "https://storage.example.com", azureContainer: "images"}

	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	fw, err := mw.CreateFormFile("image", "test.txt")
	require.NoError(t, err)
	fmt.Fprint(fw, "plain text content that is not an image")
	mw.Close()

	r := httptest.NewRequest(http.MethodPost, "/api/upload/image", &buf)
	r.Header.Set("Content-Type", mw.FormDataContentType())
	w := httptest.NewRecorder()

	srv.HandleUploadImage(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
