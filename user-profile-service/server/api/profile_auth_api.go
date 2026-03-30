package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/bankease/user-profile-service/server/db"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi/v5"
)

//#region GetMyProfile

// HandleGetMyProfile handles GET /api/profile
func (s *Server) HandleGetMyProfile(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		writeError(w, http.StatusUnauthorized, "Authorization header is required")
		return
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	userID, err := parseJWTUserID(tokenStr, s.jwtSecret)
	if err != nil {
		log.Printf("JWT parse error: %v", err)
		writeError(w, http.StatusUnauthorized, "Invalid or expired token")
		return
	}

	profile, err := s.provider.GetProfileByUserID(r.Context(), userID)
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

//#endregion

//#region GetProfile

// HandleGetProfile handles GET /api/profile/{id}
func (s *Server) HandleGetProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Profile ID is required")
		return
	}

	profile, err := s.provider.GetProfileByID(r.Context(), id)
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

//#endregion

//#region UpdateProfile

// HandleUpdateProfile handles PUT /api/profile/{id}
func (s *Server) HandleUpdateProfile(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		writeError(w, http.StatusBadRequest, "Profile ID is required")
		return
	}

	var req db.EditProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := s.provider.UpdateProfile(r.Context(), id, req); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "Profile not found")
			return
		}
		log.Printf("Error updating profile: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, StandardResponse{
		Code:        http.StatusOK,
		Description: "Data pengguna berhasil diubah.",
	})
}

//#endregion

//#region GetProfileByUserID

// HandleGetProfileByUserID handles GET /api/profile/user/{user_id}
func (s *Server) HandleGetProfileByUserID(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "user_id")
	if userID == "" {
		writeError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	profile, err := s.provider.GetProfileByUserID(r.Context(), userID)
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

//#endregion

//#region CreateProfile

// HandleCreateProfile handles POST /api/profile
func (s *Server) HandleCreateProfile(w http.ResponseWriter, r *http.Request) {
	var req db.CreateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.UserID == "" {
		writeError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	profile, err := s.provider.CreateProfile(r.Context(), req)
	if err != nil {
		log.Printf("Error creating profile: %v", err)
		writeError(w, http.StatusInternalServerError, "Failed to create profile")
		return
	}

	writeJSON(w, http.StatusCreated, profile)
}

//#endregion

//#region JWT helper

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

//#endregion
