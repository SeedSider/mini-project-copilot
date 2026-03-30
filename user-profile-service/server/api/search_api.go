package api

import (
	"log"
	"net/http"
)

// HandleGetExchangeRates handles GET /api/exchange-rates
func (s *Server) HandleGetExchangeRates(w http.ResponseWriter, r *http.Request) {
	rates, err := s.provider.GetAllExchangeRates(r.Context())
	if err != nil {
		log.Printf("Error getting exchange rates: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, rates)
}

// HandleGetInterestRates handles GET /api/interest-rates
func (s *Server) HandleGetInterestRates(w http.ResponseWriter, r *http.Request) {
	rates, err := s.provider.GetAllInterestRates(r.Context())
	if err != nil {
		log.Printf("Error getting interest rates: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, rates)
}

// HandleGetBranches handles GET /api/branches?q={query}
func (s *Server) HandleGetBranches(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	var err error
	var branches interface{}

	if q == "" {
		branches, err = s.provider.GetAllBranches(r.Context())
	} else {
		branches, err = s.provider.SearchBranchesByName(r.Context(), q)
	}

	if err != nil {
		log.Printf("Error getting branches: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, branches)
}
