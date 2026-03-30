package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/bankease/user-profile-service/internal/models"
)

// ExchangeRateRepository handles database operations for exchange rates.
type ExchangeRateRepository struct {
	DB *sql.DB
}

// GetAll retrieves all exchange rates.
func (r *ExchangeRateRepository) GetAll(ctx context.Context) ([]models.ExchangeRate, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT id, country, currency, country_code, buy, sell
		FROM exchange_rate ORDER BY country ASC`

	rows, err := r.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.ExchangeRate
	for rows.Next() {
		var e models.ExchangeRate
		if err := rows.Scan(&e.ID, &e.Country, &e.Currency, &e.CountryCode, &e.Buy, &e.Sell); err != nil {
			return nil, err
		}
		items = append(items, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if items == nil {
		items = []models.ExchangeRate{}
	}

	return items, nil
}
