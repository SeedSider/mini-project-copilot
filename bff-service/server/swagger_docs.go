package main

import (
	bff_service "github.com/bankease/bff-service/protogen/bff-service"
)

// Ensure protogen types are available for swagger type resolution.
var (
	_ *bff_service.ErrorBodyResponse
	_ *bff_service.SignUpRequest
	_ *bff_service.ValidateOtpRequest
	_ *bff_service.ValidateOtpResponse
	_ *bff_service.UpdatePasswordRequest
	_ *bff_service.UpdatePasswordResponse
	_ *bff_service.ExchangeRateItem
	_ *bff_service.ExchangeRateListResponse
	_ *bff_service.InterestRateItem
	_ *bff_service.InterestRateListResponse
	_ *bff_service.BranchItem
	_ *bff_service.BranchListResponse
	_ *bff_service.ProviderItem
	_ *bff_service.ProviderListResponse
	_ *bff_service.InternetBillDetail
	_ *bff_service.InternetBillResponse
	_ *bff_service.CurrencyEntry
	_ *bff_service.CurrencyListResponse
	_ *bff_service.BeneficiaryItem
	_ *bff_service.BeneficiaryListResponse
	_ *bff_service.PrepaidPayRequest
	_ *bff_service.PrepaidPayResponse
	_ *bff_service.AddBeneficiaryRequest
	_ *bff_service.SearchBeneficiariesRequest
	_ *bff_service.PaymentCardItem
	_ *bff_service.PaymentCardListResponse
	_ *bff_service.CreatePaymentCardRequest
)

// Swagger documentation stubs for endpoints that share a single handler
// for multiple HTTP methods. These functions exist solely for swaggo
// annotation parsing and are never called at runtime.

// swagGetMyProfile godoc
// @Summary Get my profile
// @Description Retrieve the authenticated user's banking profile using JWT token
// @Tags Profile
// @Produce json
// @Security BearerAuth
// @Success 200 {object} bff_service.ProfileResponse
// @Failure 401 {object} bff_service.ErrorBodyResponse
// @Failure 404 {object} bff_service.ErrorBodyResponse
// @Failure 500 {object} bff_service.ErrorBodyResponse
// @Router /api/profile [get]
func swagGetMyProfile() {}

// swagCreateProfile godoc
// @Summary Create a new profile
// @Description Create a new user banking profile
// @Tags Profile
// @Accept json
// @Produce json
// @Param request body bff_service.CreateProfileRequest true "Create profile request body"
// @Success 201 {object} bff_service.ProfileResponse
// @Failure 400 {object} bff_service.ErrorBodyResponse
// @Failure 500 {object} bff_service.ErrorBodyResponse
// @Router /api/profile [post]
func swagCreateProfile() {}

// swagGetProfileByID godoc
// @Summary Get profile by ID
// @Description Retrieve a user profile by its UUID
// @Tags Profile
// @Produce json
// @Param id path string true "Profile ID (UUID)"
// @Success 200 {object} bff_service.ProfileResponse
// @Failure 400 {object} bff_service.ErrorBodyResponse
// @Failure 404 {object} bff_service.ErrorBodyResponse
// @Failure 500 {object} bff_service.ErrorBodyResponse
// @Router /api/profile/{id} [get]
func swagGetProfileByID() {}

// swagUpdateProfile godoc
// @Summary Update a profile
// @Description Update an existing user profile by ID
// @Tags Profile
// @Accept json
// @Produce json
// @Param id path string true "Profile ID (UUID)"
// @Param request body bff_service.UpdateProfileRequest true "Update profile request body"
// @Success 200 {object} bff_service.StandardResponse
// @Failure 400 {object} bff_service.ErrorBodyResponse
// @Failure 404 {object} bff_service.ErrorBodyResponse
// @Failure 500 {object} bff_service.ErrorBodyResponse
// @Router /api/profile/{id} [put]
func swagUpdateProfile() {}

// swagUploadImage godoc
// @Summary Upload image
// @Description Upload an image file to Azure Blob Storage (max 5MB). Supported formats: JPEG, PNG, GIF, WebP, SVG.
// @Tags Upload
// @Accept multipart/form-data
// @Produce json
// @Param image formData file true "Image file (max 5MB)"
// @Success 200 {object} uploadResponse
// @Failure 400 {object} bff_service.ErrorBodyResponse
// @Failure 413 {object} bff_service.ErrorBodyResponse "File too large"
// @Failure 500 {object} bff_service.ErrorBodyResponse
// @Failure 503 {object} bff_service.ErrorBodyResponse "Azure not configured"
// @Router /api/upload/image [post]
func swagUploadImage() {}
