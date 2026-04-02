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

// HandleAddBeneficiary handles POST /api/mobile-prepaid/beneficiaries
// @Summary      Add beneficiary
// @Description  Save a new mobile prepaid top-up contact
// @Tags         Mobile Prepaid
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body addBeneficiaryHTTPRequest true "Add beneficiary request body"
// @Success      201  {object}  db.Beneficiary
// @Failure      400  {object}  prepaidErrorResponse
// @Failure      401  {object}  prepaidErrorResponse
// @Failure      500  {object}  prepaidErrorResponse
// @Router       /api/mobile-prepaid/beneficiaries [post]
func (s *Server) HandleAddBeneficiary(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value("user_claims").(*manager.UserClaims)
	if !ok {
		writePrepaidError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid Bearer token")
		return
	}

	var req addBeneficiaryHTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writePrepaidError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
		return
	}

	if errMsg := validateAddBeneficiary(req); errMsg != "" {
		writePrepaidError(w, http.StatusBadRequest, "VALIDATION_ERROR", errMsg)
		return
	}

	created, err := s.provider.CreateBeneficiary(r.Context(), db.Beneficiary{
		AccountID: req.AccountID,
		Name:      req.Name,
		Phone:     req.Phone,
		Avatar:    req.Avatar,
	})
	if err != nil {
		log.Error("", "HandleAddBeneficiary", err.Error(), nil, nil, nil, err)
		writePrepaidError(w, http.StatusInternalServerError, "INTERNAL_ERROR", InternalServerErrorMessage)
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

type addBeneficiaryHTTPRequest struct {
	AccountID string `json:"accountId"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Avatar    string `json:"avatar"`
}

func validateAddBeneficiary(req addBeneficiaryHTTPRequest) string {
	if req.AccountID == "" {
		return "accountId is required"
	}
	if req.Name == "" {
		return "name is required"
	}
	if req.Phone == "" {
		return "phone is required"
	}
	if !phoneRegex.MatchString(req.Phone) {
		return "phone must be 10-13 digits"
	}
	return ""
}

// HandleSearchBeneficiaries handles GET /api/mobile-prepaid/beneficiaries/search?accountId={id}&q={query}
// @Summary      Search beneficiaries
// @Description  Search saved mobile prepaid contacts by name or phone
// @Tags         Mobile Prepaid
// @Produce      json
// @Security     BearerAuth
// @Param        accountId query string true "Account ID"
// @Param        q query string true "Search query"
// @Success      200  {array}   db.Beneficiary
// @Failure      401  {object}  prepaidErrorResponse
// @Failure      500  {object}  prepaidErrorResponse
// @Router       /api/mobile-prepaid/beneficiaries/search [get]
func (s *Server) HandleSearchBeneficiaries(w http.ResponseWriter, r *http.Request) {
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

	query := r.URL.Query().Get("q")
	if query == "" {
		writePrepaidError(w, http.StatusBadRequest, "VALIDATION_ERROR", "q query parameter is required")
		return
	}

	beneficiaries, err := s.provider.SearchBeneficiaries(r.Context(), accountID, query)
	if err != nil {
		log.Error("", "HandleSearchBeneficiaries", err.Error(), nil, nil, nil, err)
		writePrepaidError(w, http.StatusInternalServerError, "INTERNAL_ERROR", InternalServerErrorMessage)
		return
	}

	if beneficiaries == nil {
		beneficiaries = []db.Beneficiary{}
	}
	writeJSON(w, http.StatusOK, beneficiaries)
}

// HandleGetPaymentCards handles GET /api/mobile-prepaid/cards?accountId={id}
// @Summary      Get payment cards
// @Description  Retrieve all payment cards for an account
// @Tags         Mobile Prepaid
// @Produce      json
// @Security     BearerAuth
// @Param        accountId query string true "Account ID"
// @Success      200  {array}   db.PaymentCard
// @Failure      401  {object}  prepaidErrorResponse
// @Failure      500  {object}  prepaidErrorResponse
// @Router       /api/mobile-prepaid/cards [get]
func (s *Server) HandleGetPaymentCards(w http.ResponseWriter, r *http.Request) {
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

	cards, err := s.provider.GetCardsByAccountID(r.Context(), accountID)
	if err != nil {
		log.Error("", "HandleGetPaymentCards", err.Error(), nil, nil, nil, err)
		writePrepaidError(w, http.StatusInternalServerError, "INTERNAL_ERROR", InternalServerErrorMessage)
		return
	}

	if cards == nil {
		cards = []db.PaymentCard{}
	}
	writeJSON(w, http.StatusOK, cards)
}

// HandleCreatePaymentCard handles POST /api/mobile-prepaid/cards
// @Summary      Create payment card
// @Description  Register a new payment card/account
// @Tags         Mobile Prepaid
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body createCardHTTPRequest true "Create card request body"
// @Success      201  {object}  db.PaymentCard
// @Failure      400  {object}  prepaidErrorResponse
// @Failure      401  {object}  prepaidErrorResponse
// @Failure      500  {object}  prepaidErrorResponse
// @Router       /api/mobile-prepaid/cards [post]
func (s *Server) HandleCreatePaymentCard(w http.ResponseWriter, r *http.Request) {
	_, ok := r.Context().Value("user_claims").(*manager.UserClaims)
	if !ok {
		writePrepaidError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Missing or invalid Bearer token")
		return
	}

	var req createCardHTTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writePrepaidError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid request body")
		return
	}

	if errMsg := validateCreateCard(req); errMsg != "" {
		writePrepaidError(w, http.StatusBadRequest, "VALIDATION_ERROR", errMsg)
		return
	}

	created, err := s.provider.CreatePaymentCard(r.Context(), db.PaymentCard{
		AccountID:      req.AccountID,
		HolderName:     req.HolderName,
		CardLabel:      req.CardLabel,
		MaskedNumber:   req.MaskedNumber,
		Balance:        req.Balance,
		Currency:       req.Currency,
		Brand:          req.Brand,
		GradientColors: req.GradientColors,
	})
	if err != nil {
		log.Error("", "HandleCreatePaymentCard", err.Error(), nil, nil, nil, err)
		writePrepaidError(w, http.StatusInternalServerError, "INTERNAL_ERROR", InternalServerErrorMessage)
		return
	}

	writeJSON(w, http.StatusCreated, created)
}

type createCardHTTPRequest struct {
	AccountID      string   `json:"accountId"`
	HolderName     string   `json:"holderName"`
	CardLabel      string   `json:"cardLabel"`
	MaskedNumber   string   `json:"maskedNumber"`
	Balance        int64    `json:"balance"`
	Currency       string   `json:"currency"`
	Brand          string   `json:"brand"`
	GradientColors []string `json:"gradientColors"`
}

func validateCreateCard(req createCardHTTPRequest) string {
	if req.AccountID == "" {
		return "accountId is required"
	}
	if req.HolderName == "" {
		return "holderName is required"
	}
	return ""
}
