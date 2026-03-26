package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/bankease/user-profile-service/internal/models"
	"github.com/bankease/user-profile-service/internal/repository"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi/v5"
)

// ProfileHandler handles HTTP requests for profile endpoints.
type ProfileHandler struct {
	Repo      *repository.ProfileRepository
	JWTSecret string
}

// GetMyProfile handles GET /api/profile
// @Summary      Get my profile
// @Description  Retrieve the authenticated user's banking profile using JWT token
// @Tags         Profile
// @Produce      json
// @Security     BearerAuth
// @Param        Authorization  header  string  true  "Bearer <token>"
// @Success      200  {object}  models.Profile
// @Failure      401  {object}  models.StandardResponse
// @Failure      404  {object}  models.StandardResponse
// @Failure      500  {object}  models.StandardResponse
// @Router       /api/profile [get]
func (h *ProfileHandler) GetMyProfile(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		writeError(w, http.StatusUnauthorized, "Authorization header is required")
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	userID, err := parseJWTUserID(tokenStr, h.JWTSecret)
	if err != nil {
		log.Printf("JWT parse error: %v", err)
		writeError(w, http.StatusUnauthorized, "Invalid or expired token")
		return
	}

	profile, err := h.Repo.GetProfileByUserID(r.Context(), userID)
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

// GetProfile handles GET /api/profile/{id}
// @Summary      Get user profile
// @Description  Retrieve a user profile by ID
// @Tags         Profile
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Profile ID (UUID)"
// @Success      200  {object}  models.Profile
// @Failure      400  {object}  models.StandardResponse
// @Failure      404  {object}  models.StandardResponse
// @Failure      500  {object}  models.StandardResponse
// @Router       /api/profile/{id} [get]
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
// @Summary      Update user profile
// @Description  Update editable fields of a user profile
// @Tags         Profile
// @Accept       json
// @Produce      json
// @Param        id    path      string                    true  "Profile ID (UUID)"
// @Param        body  body      models.EditProfileRequest  true  "Profile fields to update"
// @Success      200   {object}  models.StandardResponse
// @Failure      400   {object}  models.StandardResponse
// @Failure      404   {object}  models.StandardResponse
// @Failure      500   {object}  models.StandardResponse
// @Router       /api/profile/{id} [put]
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

// GetProfileByUserID handles GET /api/profile/user/{user_id}
func (h *ProfileHandler) GetProfileByUserID(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	profile, err := h.Repo.GetProfileByUserID(r.Context(), userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Profile not found")
			return
		}
		log.Printf("Error getting profile by user_id: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, profile)
}

// CreateProfile handles POST /api/profile
// @Summary      Create a new user profile
// @Description  Create a new banking profile linked to a user ID
// @Tags         Profile
// @Accept       json
// @Produce      json
// @Param        body  body      models.CreateProfileRequest  true  "Profile data"
// @Success      201   {object}  models.Profile
// @Failure      400   {object}  models.StandardResponse
// @Failure      500   {object}  models.StandardResponse
// @Router       /api/profile [post]
func (h *ProfileHandler) CreateProfile(w http.ResponseWriter, r *http.Request) {
	var req models.CreateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.UserID == "" {
		writeError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	profile, err := h.Repo.CreateProfile(r.Context(), req)
	if err != nil {
		log.Printf("Error creating profile: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to create profile")
		return
	}

	writeJSON(w, http.StatusCreated, profile)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, models.StandardResponse{Code: status, Description: message})
}

func parseJWTUserID(tokenStr, secret string) (string, error) {
	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid claims")
	}

	userID, ok := claims["user_id"].(string)
	if !ok || userID == "" {
		return "", fmt.Errorf("user_id not found in token")
	}

	return userID, nil
}
