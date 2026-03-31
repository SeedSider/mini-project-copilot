package api

import (
	"encoding/json"
	"net/http"
)

// HandleGetExchangeRates handles GET /api/exchange-rates
// @Summary      Get all exchange rates
// @Description  Retrieve all currency exchange rates
// @Tags         Exchange Rates
// @Produce      json
// @Success      200  {array}   db.ExchangeRate
// @Failure      500  {object}  standardResponse
// @Router       /api/exchange-rates [get]
func (s *Server) HandleGetExchangeRates(w http.ResponseWriter, r *http.Request) {
	rates, err := s.provider.GetAllExchangeRates(r.Context())
	if err != nil {
		log.Error("", "HandleGetExchangeRates", err.Error(), nil, nil, nil, err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, rates)
}

// HandleGetInterestRates handles GET /api/interest-rates
// @Summary      Get all interest rates
// @Description  Retrieve all deposit interest rates
// @Tags         Interest Rates
// @Produce      json
// @Success      200  {array}   db.InterestRate
// @Failure      500  {object}  standardResponse
// @Router       /api/interest-rates [get]
func (s *Server) HandleGetInterestRates(w http.ResponseWriter, r *http.Request) {
	rates, err := s.provider.GetAllInterestRates(r.Context())
	if err != nil {
		log.Error("", "HandleGetInterestRates", err.Error(), nil, nil, nil, err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, rates)
}

// HandleGetBranches handles GET /api/branches?q={query}
// @Summary      Get branches
// @Description  Retrieve all branches, or search by name with query parameter
// @Tags         Branches
// @Produce      json
// @Param        q  query  string  false  "Search query (case-insensitive partial match on branch name)"
// @Success      200  {array}   db.Branch
// @Failure      500  {object}  standardResponse
// @Router       /api/branches [get]
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
		log.Error("", "HandleGetBranches", err.Error(), nil, nil, nil, err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, branches)
}

type standardResponse struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, standardResponse{Code: status, Description: message})
}
