package handlers

import (
	"log"
	"net/http"

	"github.com/bankease/user-profile-service/internal/repository"
)

// SearchHandler handles HTTP requests for search/rates/branches endpoints.
type SearchHandler struct {
	ExchangeRateRepo *repository.ExchangeRateRepository
	InterestRateRepo *repository.InterestRateRepository
	BranchRepo       *repository.BranchRepository
}

// GetExchangeRates handles GET /api/exchange-rates
// @Summary      Get exchange rates
// @Description  Retrieve a list of currency exchange rates
// @Tags         Search
// @Produce      json
// @Success      200  {array}   models.ExchangeRate
// @Failure      500  {object}  models.StandardResponse
// @Router       /api/exchange-rates [get]
func (h *SearchHandler) GetExchangeRates(w http.ResponseWriter, r *http.Request) {
	rates, err := h.ExchangeRateRepo.GetAll(r.Context())
	if err != nil {
		log.Printf("Error getting exchange rates: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, rates)
}

// GetInterestRates handles GET /api/interest-rates
// @Summary      Get interest rates
// @Description  Retrieve a list of deposit interest rates
// @Tags         Search
// @Produce      json
// @Success      200  {array}   models.InterestRate
// @Failure      500  {object}  models.StandardResponse
// @Router       /api/interest-rates [get]
func (h *SearchHandler) GetInterestRates(w http.ResponseWriter, r *http.Request) {
	rates, err := h.InterestRateRepo.GetAll(r.Context())
	if err != nil {
		log.Printf("Error getting interest rates: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, rates)
}

// GetBranches handles GET /api/branches?q={query}
// Returns all branches when q is empty; otherwise filters by name (case-insensitive partial match).
// @Summary      Get branches
// @Description  Retrieve a list of BRI branches. Optionally filter by name using the q query parameter (case-insensitive partial match).
// @Tags         Search
// @Produce      json
// @Param        q  query  string  false  "Search query — filters by branch name"
// @Success      200  {array}   models.Branch
// @Failure      500  {object}  models.StandardResponse
// @Router       /api/branches [get]
func (h *SearchHandler) GetBranches(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")

	var err error
	var branches interface{}

	if q == "" {
		branches, err = h.BranchRepo.GetAll(r.Context())
	} else {
		branches, err = h.BranchRepo.SearchByName(r.Context(), q)
	}

	if err != nil {
		log.Printf("Error getting branches: %v", err)
		writeError(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	writeJSON(w, http.StatusOK, branches)
}
