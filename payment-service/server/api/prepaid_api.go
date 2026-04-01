package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"

	manager "github.com/bankease/payment-service/server/jwt"

	"github.com/bankease/payment-service/server/db"
)

var phoneRegex = regexp.MustCompile(`^\d{10,13}$`)

// HandleGetBeneficiaries handles GET /api/mobile-prepaid/beneficiaries?accountId={id}
// @Summary      Get beneficiaries
// @Description  Retrieve saved mobile prepaid top-up contacts for an account
// @Tags         Mobile Prepaid
// @Produce      json
// @Security     BearerAuth
// @Param        accountId query string true "Account ID"
// @Success      200  {array}   db.Beneficiary
// @Failure      401  {object}  prepaidErrorResponse
// @Failure      404  {object}  prepaidErrorResponse
// @Failure      500  {object}  prepaidErrorResponse
// @Router       /api/mobile-prepaid/beneficiaries [get]
func (s *Server) HandleGetBeneficiaries(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value("user_claims").(*manager.UserClaims)
	if !ok {
		writePrepaidError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid Bearer token")
		return
	}

	accountID := r.URL.Query().Get("accountId")
	if accountID == "" {
		writePrepaidError(w, http.StatusBadRequest, "VALIDATION_ERROR", "accountId query parameter is required")
		return
	}

	beneficiaries, err := s.provider.GetBeneficiariesByAccountID(r.Context(), accountID)
	if err != nil {
		log.Error("", "HandleGetBeneficiaries", err.Error(), nil, nil, nil, err)
		writePrepaidError(w, http.StatusInternalServerError, "INTERNAL_ERROR", InternalServerErrorMessage)
		return
	}

	if beneficiaries == nil {
		beneficiaries = []db.Beneficiary{}
	}
	writeJSON(w, http.StatusOK, beneficiaries)
}

// HandlePrepaidPay handles POST /api/mobile-prepaid/pay
// @Summary      Submit prepaid payment
// @Description  Process a mobile prepaid top-up
// @Tags         Mobile Prepaid
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        Idempotency-Key header string true "Client-generated UUID"
// @Param        request body prepaidPayHTTPRequest true "Payment request body"
// @Success      200  {object}  db.PrepaidTransaction
// @Failure      400  {object}  prepaidErrorResponse
// @Failure      401  {object}  prepaidErrorResponse
// @Failure      409  {object}  prepaidErrorResponse
// @Failure      500  {object}  prepaidErrorResponse
// @Router       /api/mobile-prepaid/pay [post]
func (s *Server) HandlePrepaidPay(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value("user_claims").(*manager.UserClaims)
	if !ok {
		writePrepaidError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid Bearer token")
		return
	}

	idempotencyKey := r.Header.Get("Idempotency-Key")
	if idempotencyKey == "" {
		writePrepaidError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Idempotency-Key header is required")
		return
	}

	var req prepaidPayHTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writePrepaidError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
		return
	}

	if errMsg := validatePrepaidPay(req); errMsg != "" {
		writePrepaidError(w, http.StatusBadRequest, "VALIDATION_ERROR", errMsg)
		return
	}

	// Check idempotency: if the key already exists, return the original response
	existing, err := s.provider.GetTransactionByIdempotencyKey(r.Context(), idempotencyKey)
	if err != nil {
		log.Error("", "HandlePrepaidPay", err.Error(), nil, nil, nil, err)
		writePrepaidError(w, http.StatusInternalServerError, "INTERNAL_ERROR", InternalServerErrorMessage)
		return
	}
	if existing != nil {
		writeJSON(w, http.StatusOK, existing)
		return
	}

	txnID := fmt.Sprintf("txn-%d", time.Now().UnixMilli())
	txn := db.PrepaidTransaction{
		ID:             txnID,
		CardID:         req.CardID,
		Phone:          req.Phone,
		Amount:         req.Amount,
		Status:         "SUCCESS",
		Message:        "Top-up successful",
		IdempotencyKey: idempotencyKey,
	}

	result, err := s.provider.CreatePrepaidTransaction(r.Context(), txn)
	if err != nil {
		log.Error("", "HandlePrepaidPay", err.Error(), nil, nil, nil, err)
		writePrepaidError(w, http.StatusInternalServerError, "INTERNAL_ERROR", InternalServerErrorMessage)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

type prepaidPayHTTPRequest struct {
	CardID string `json:"cardId"`
	Phone  string `json:"phone"`
	Amount int64  `json:"amount"`
}

func validatePrepaidPay(req prepaidPayHTTPRequest) string {
	if req.CardID == "" {
		return "cardId is required"
	}
	if req.Phone == "" {
		return "phone is required"
	}
	if !phoneRegex.MatchString(req.Phone) {
		return "phone must be 10-13 digits"
	}
	if req.Amount <= 0 {
		return "amount must be greater than 0"
	}
	return ""
}
