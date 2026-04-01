package api

import (
	"net/http"

	manager "github.com/bankease/payment-service/server/jwt"
)

const InternalServerErrorMessage = "Internal server error"

// HandleGetProviders handles GET /api/pay-the-bill/providers
// @Summary      Get all providers
// @Description  Retrieve all internet service providers
// @Tags         Pay The Bill
// @Produce      json
// @Success      200  {array}   db.ServiceProvider
// @Failure      500  {object}  standardResponse
// @Router       /api/pay-the-bill/providers [get]
func (s *Server) HandleGetProviders(w http.ResponseWriter, r *http.Request) {
	providers, err := s.provider.GetAllProviders(r.Context())
	if err != nil {
		log.Error("", "HandleGetProviders", err.Error(), nil, nil, nil, err)
		writeError(w, http.StatusInternalServerError, InternalServerErrorMessage)
		return
	}

	writeJSON(w, http.StatusOK, providers)
}

// HandleGetInternetBill handles GET /api/pay-the-bill/internet-bill (JWT protected)
// @Summary      Get internet bill
// @Description  Retrieve internet bill for the authenticated user
// @Tags         Pay The Bill
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  db.InternetBill
// @Failure      401  {object}  errorResponse
// @Failure      404  {object}  standardResponse
// @Failure      500  {object}  standardResponse
// @Router       /api/pay-the-bill/internet-bill [get]
func (s *Server) HandleGetInternetBill(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value("user_claims").(*manager.UserClaims)
	if !ok || claims == nil {
		writeAuthError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	bill, err := s.provider.GetInternetBillByUserID(r.Context(), claims.UserID)
	if err != nil {
		log.Error("", "HandleGetInternetBill", err.Error(), nil, nil, nil, err)
		writeError(w, http.StatusNotFound, "Internet bill not found")
		return
	}

	writeJSON(w, http.StatusOK, bill)
}

// HandleGetCurrencyList handles GET /api/currency-list
// @Summary      Get currency list
// @Description  Retrieve all supported currencies with exchange rates
// @Tags         Currency
// @Produce      json
// @Success      200  {array}   db.Currency
// @Failure      500  {object}  standardResponse
// @Router       /api/currency-list [get]
func (s *Server) HandleGetCurrencyList(w http.ResponseWriter, r *http.Request) {
	currencies, err := s.provider.GetAllCurrencies(r.Context())
	if err != nil {
		log.Error("", "HandleGetCurrencyList", err.Error(), nil, nil, nil, err)
		writeError(w, http.StatusInternalServerError, InternalServerErrorMessage)
		return
	}

	writeJSON(w, http.StatusOK, currencies)
}
