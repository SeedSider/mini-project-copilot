package models

// Profile represents a user profile record in the database.
type Profile struct {
	ID           string  `json:"id"`
	UserID       *string `json:"user_id,omitempty"`
	Bank         string  `json:"bank"`
	Branch       string  `json:"branch"`
	Name         string  `json:"name"`
	CardNumber   string  `json:"card_number"`
	CardProvider string  `json:"card_provider"`
	Balance      int64   `json:"balance"`
	Currency     string  `json:"currency"`
	AccountType  string  `json:"accountType"`
	Image        string  `json:"image"`
}

// EditProfileRequest contains the fields that can be updated via PUT.
type EditProfileRequest struct {
	Bank       string `json:"bank"`
	Branch     string `json:"branch"`
	Name       string `json:"name"`
	CardNumber string `json:"card_number"`
}

// CreateProfileRequest contains the fields for creating a new profile via POST.
type CreateProfileRequest struct {
	UserID       string `json:"user_id"`
	Bank         string `json:"bank"`
	Branch       string `json:"branch"`
	Name         string `json:"name"`
	CardNumber   string `json:"card_number"`
	CardProvider string `json:"card_provider"`
	Balance      int64  `json:"balance"`
	Currency     string `json:"currency"`
	AccountType  string `json:"accountType"`
}

// StandardResponse is the consistent response format for success/error.
type StandardResponse struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

// UploadResponse is returned after a successful image upload.
type UploadResponse struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
	URL         string `json:"url"`
}
