package handlers

import (
	"log"
	"net/http"

	"github.com/bankease/user-profile-service/internal/models"
	"github.com/bankease/user-profile-service/internal/repository"
	"github.com/go-chi/chi/v5"
)

// MenuHandler handles HTTP requests for menu endpoints.
type MenuHandler struct {
	Repo *repository.MenuRepository
}

// GetAllMenus handles GET /api/menu
func (h *MenuHandler) GetAllMenus(w http.ResponseWriter, r *http.Request) {
	menus, err := h.Repo.GetAllMenus(r.Context())
	if err != nil {
		log.Printf("Error getting menus: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, models.MenuResponse{Menus: menus})
}

// GetMenusByAccountType handles GET /api/menu/{accountType}
func (h *MenuHandler) GetMenusByAccountType(w http.ResponseWriter, r *http.Request) {
	accountType := chi.URLParam(r, "accountType")
	if accountType == "" {
		writeError(w, http.StatusBadRequest, "Account type is required")
		return
	}

	menus, err := h.Repo.GetMenusByAccountType(r.Context(), accountType)
	if err != nil {
		log.Printf("Error getting menus by account type: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, models.MenuResponse{Menus: menus})
}
