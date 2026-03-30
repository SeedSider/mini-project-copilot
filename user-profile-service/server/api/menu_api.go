package api

import (
	"log"
	"net/http"

	"github.com/bankease/user-profile-service/server/db"
	"github.com/go-chi/chi/v5"
)

// HandleGetAllMenus handles GET /api/menu
func (s *Server) HandleGetAllMenus(w http.ResponseWriter, r *http.Request) {
	menus, err := s.provider.GetAllMenus(r.Context())
	if err != nil {
		log.Printf("Error getting menus: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, db.MenuResponse{Menus: menus})
}

// HandleGetMenusByAccountType handles GET /api/menu/{accountType}
func (s *Server) HandleGetMenusByAccountType(w http.ResponseWriter, r *http.Request) {
	accountType := chi.URLParam(r, "accountType")
	if accountType == "" {
		writeError(w, http.StatusBadRequest, "Account type is required")
		return
	}

	menus, err := s.provider.GetMenusByAccountType(r.Context(), accountType)
	if err != nil {
		log.Printf("Error getting menus by account type: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, db.MenuResponse{Menus: menus})
}
