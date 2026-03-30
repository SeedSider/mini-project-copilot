package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/bankease/user-profile-service/internal/models"
)

// InterestRateRepository handles database operations for interest rates.
type InterestRateRepository struct {
	DB *sql.DB
}

// GetAll retrieves all interest rates.
func (r *InterestRateRepository) GetAll(ctx context.Context) ([]models.InterestRate, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT id, kind, deposit, rate
		FROM interest_rate ORDER BY kind ASC, deposit ASC`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.InterestRate
	for rows.Next() {
		var ir models.InterestRate
		if err := rows.Scan(&ir.ID, &ir.Kind, &ir.Deposit, &ir.Rate); err != nil {
			return nil, err
		}
		items = append(items, ir)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if items == nil {
		items = []models.InterestRate{}
	}

	return items, nil
}
