package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/bankease/user-profile-service/internal/models"
)

// ProfileRepository handles database operations for profile.
// Pattern from: addons-issuance-lc-service/server/db/issued_lc_provider.go
type ProfileRepository struct {
	DB *sql.DB
}

// GetProfileByID retrieves a profile by its UUID.
func (r *ProfileRepository) GetProfileByID(ctx context.Context, id string) (*models.Profile, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT id, user_id, bank, branch, name, card_number, card_provider, balance, currency, account_type, image
		FROM profile WHERE id = $1`

	var p models.Profile
	err := r.DB.QueryRowContext(ctx, query, id).Scan(
		&p.ID, &p.UserID, &p.Bank, &p.Branch, &p.Name,
		&p.CardNumber, &p.CardProvider, &p.Balance,
		&p.Currency, &p.AccountType, &p.Image,
	)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

// UpdateProfile updates editable fields of a profile by ID.
// Returns sql.ErrNoRows if profile not found.
func (r *ProfileRepository) UpdateProfile(ctx context.Context, id string, req models.EditProfileRequest) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `UPDATE profile
		SET bank = $1, branch = $2, name = $3, card_number = $4
		WHERE id = $5`

	result, err := r.DB.ExecContext(ctx, query,
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
func (r *ProfileRepository) GetProfileByUserID(ctx context.Context, userID string) (*models.Profile, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT id, user_id, bank, branch, name, card_number, card_provider, balance, currency, account_type, image
		FROM profile WHERE user_id = $1`

	var p models.Profile
	err := r.DB.QueryRowContext(ctx, query, userID).Scan(
		&p.ID, &p.UserID, &p.Bank, &p.Branch, &p.Name,
		&p.CardNumber, &p.CardProvider, &p.Balance,
		&p.Currency, &p.AccountType, &p.Image,
	)
	if err != nil {
		return nil, err
	}

	return &p, nil
}

// CreateProfile inserts a new profile record.
func (r *ProfileRepository) CreateProfile(ctx context.Context, req models.CreateProfileRequest) (*models.Profile, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `INSERT INTO profile (user_id, bank, branch, name, card_number, card_provider, balance, currency, account_type)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, user_id, bank, branch, name, card_number, card_provider, balance, currency, account_type, image`

	var p models.Profile
	err := r.DB.QueryRowContext(ctx, query,
		req.UserID, req.Bank, req.Branch, req.Name,
		req.CardNumber, req.CardProvider, req.Balance,
		req.Currency, req.AccountType,
	).Scan(
		&p.ID, &p.UserID, &p.Bank, &p.Branch, &p.Name,
		&p.CardNumber, &p.CardProvider, &p.Balance,
		&p.Currency, &p.AccountType, &p.Image,
	)
	if err != nil {
		return nil, err
	}

	return &p, nil
}
