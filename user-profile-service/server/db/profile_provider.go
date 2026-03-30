package db

import (
	"context"
	"database/sql"
	"time"
)

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
	Image        string `json:"image"`
}

// GetProfileByID retrieves a profile by its UUID.
func (p *Provider) GetProfileByID(ctx context.Context, id string) (*Profile, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT id, user_id, bank, branch, name, card_number, card_provider, balance, currency, account_type, image
		FROM profile WHERE id = $1`

	var profile Profile
	err := p.DB.QueryRowContext(ctx, query, id).Scan(
		&profile.ID, &profile.UserID, &profile.Bank, &profile.Branch, &profile.Name,
		&profile.CardNumber, &profile.CardProvider, &profile.Balance,
		&profile.Currency, &profile.AccountType, &profile.Image,
	)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

// UpdateProfile updates editable fields of a profile by ID.
// Returns sql.ErrNoRows if profile not found.
func (p *Provider) UpdateProfile(ctx context.Context, id string, req EditProfileRequest) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `UPDATE profile
		SET bank = $1, branch = $2, name = $3, card_number = $4
		WHERE id = $5`

	result, err := p.DB.ExecContext(ctx, query,
		req.Bank, req.Branch, req.Name,
		req.CardNumber,
		id,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

// GetProfileByUserID retrieves a profile by its linked user UUID.
func (p *Provider) GetProfileByUserID(ctx context.Context, userID string) (*Profile, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT id, user_id, bank, branch, name, card_number, card_provider, balance, currency, account_type, image
		FROM profile WHERE user_id = $1`

	var profile Profile
	err := p.DB.QueryRowContext(ctx, query, userID).Scan(
		&profile.ID, &profile.UserID, &profile.Bank, &profile.Branch, &profile.Name,
		&profile.CardNumber, &profile.CardProvider, &profile.Balance,
		&profile.Currency, &profile.AccountType, &profile.Image,
	)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

// CreateProfile inserts a new profile record.
func (p *Provider) CreateProfile(ctx context.Context, req CreateProfileRequest) (*Profile, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `INSERT INTO profile (user_id, bank, branch, name, card_number, card_provider, balance, currency, account_type, image)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, user_id, bank, branch, name, card_number, card_provider, balance, currency, account_type, image`

	var profile Profile
	err := p.DB.QueryRowContext(ctx, query,
		req.UserID, req.Bank, req.Branch, req.Name,
		req.CardNumber, req.CardProvider, req.Balance,
		req.Currency, req.AccountType, req.Image,
	).Scan(
		&profile.ID, &profile.UserID, &profile.Bank, &profile.Branch, &profile.Name,
		&profile.CardNumber, &profile.CardProvider, &profile.Balance,
		&profile.Currency, &profile.AccountType, &profile.Image,
	)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}
