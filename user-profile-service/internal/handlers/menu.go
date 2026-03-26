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
// @Summary      Get all menus
// @Description  Retrieve all homepage menu items
// @Tags         Menu
// @Produce      json
// @Success      200  {object}  models.MenuResponse
// @Failure      500  {object}  models.StandardResponse
// @Router       /api/menu [get]
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
// @Summary      Get menus by account type
// @Description  Retrieve menu items filtered by account type. PREMIUM returns all menus, REGULAR returns only REGULAR menus.
// @Tags         Menu
// @Produce      json
// @Param        accountType  path      string  true  "Account type (REGULAR or PREMIUM)"
// @Success      200          {object}  models.MenuResponse
// @Failure      400          {object}  models.StandardResponse
// @Failure      500          {object}  models.StandardResponse
// @Router       /api/menu/{accountType} [get]
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
