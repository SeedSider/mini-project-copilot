package db

import (
	"context"
	"database/sql"
	"time"
)

// ExchangeRate represents a currency exchange rate record.
type ExchangeRate struct {
	ID          string  `json:"id"`
	Country     string  `json:"country"`
	Currency    string  `json:"currency"`
	CountryCode string  `json:"countryCode"`
	Buy         float64 `json:"buy"`
	Sell        float64 `json:"sell"`
}

// InterestRate represents a deposit interest rate record.
type InterestRate struct {
	ID      string  `json:"id"`
	Kind    string  `json:"kind"`
	Deposit string  `json:"deposit"`
	Rate    float64 `json:"rate"`
}

// Branch represents a bank branch record.
type Branch struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Distance  string  `json:"distance"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// GetAllExchangeRates retrieves all exchange rates.
func (p *Provider) GetAllExchangeRates(ctx context.Context) ([]ExchangeRate, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT id, country, currency, country_code, buy, sell
		FROM exchange_rate ORDER BY country ASC`

	rows, err := p.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []ExchangeRate
	for rows.Next() {
		var e ExchangeRate
		if err := rows.Scan(&e.ID, &e.Country, &e.Currency, &e.CountryCode, &e.Buy, &e.Sell); err != nil {
			return nil, err
		}
		items = append(items, e)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if items == nil {
		items = []ExchangeRate{}
	}

	return items, nil
}

// GetAllInterestRates retrieves all interest rates.
func (p *Provider) GetAllInterestRates(ctx context.Context) ([]InterestRate, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT id, kind, deposit, rate
		FROM interest_rate ORDER BY kind ASC, deposit ASC`

	rows, err := p.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []InterestRate
	for rows.Next() {
		var ir InterestRate
		if err := rows.Scan(&ir.ID, &ir.Kind, &ir.Deposit, &ir.Rate); err != nil {
			return nil, err
		}
		items = append(items, ir)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if items == nil {
		items = []InterestRate{}
	}

	return items, nil
}

// GetAllBranches retrieves all branches.
func (p *Provider) GetAllBranches(ctx context.Context) ([]Branch, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT id, name, distance, latitude, longitude FROM branch`
	return p.scanBranches(ctx, query)
}

// SearchBranchesByName retrieves branches whose name matches the query (case-insensitive partial match).
func (p *Provider) SearchBranchesByName(ctx context.Context, q string) ([]Branch, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query := `SELECT id, name, distance, latitude, longitude
		FROM branch WHERE name ILIKE '%' || $1 || '%'`
	return p.scanBranchesWithParam(ctx, query, q)
}

func (p *Provider) scanBranches(ctx context.Context, query string) ([]Branch, error) {
	rows, err := p.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanBranchRows(rows)
}

func (p *Provider) scanBranchesWithParam(ctx context.Context, query string, param string) ([]Branch, error) {
	rows, err := p.DB.QueryContext(ctx, query, param)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanBranchRows(rows)
}

func scanBranchRows(rows *sql.Rows) ([]Branch, error) {
	var items []Branch
	for rows.Next() {
		var b Branch
		if err := rows.Scan(&b.ID, &b.Name, &b.Distance, &b.Latitude, &b.Longitude); err != nil {
			return nil, err
		}
		items = append(items, b)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if items == nil {
		items = []Branch{}
	}

	return items, nil
}
