package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/bankease/user-profile-service/internal/models"
	"github.com/bankease/user-profile-service/internal/repository"
	"github.com/go-chi/chi/v5"
)

// ProfileHandler handles HTTP requests for profile endpoints.
// Pattern from: addons-issuance-lc-service/server/api/issued_lc_data_api.go
type ProfileHandler struct {
	Repo *repository.ProfileRepository
}

// GetProfile handles GET /api/profile/{id}
func (h *ProfileHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Profile ID is required")
		return
	}

	profile, err := h.Repo.GetProfileByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Profile not found")
			return
		}
		log.Printf("Error getting profile: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, profile)
}

// UpdateProfile handles PUT /api/profile/{id}
func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Profile ID is required")
		return
	}

	var req models.EditProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate currency
	if req.Currency != "IDR" && req.Currency != "USD" {
		writeError(w, http.StatusBadRequest, "Currency tidak didukung")
		return
	}

	if err := h.Repo.UpdateProfile(r.Context(), id, req); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Profile not found")
			return
		}
		log.Printf("Error updating profile: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, models.StandardResponse{
		Code:        http.StatusOK,
		Description: "Data pengguna berhasil diubah.",
	})
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, description string) {
	writeJSON(w, status, models.StandardResponse{
		Code:        status,
		Description: description,
	})
}
