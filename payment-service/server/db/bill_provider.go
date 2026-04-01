package db

import (
	"context"
	"database/sql"
)

// ServiceProvider represents a bill provider (e.g. Biznet, Indihome).
type ServiceProvider struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// InternetBill represents a customer's internet bill detail.
type InternetBill struct {
	ID          string `json:"-"`
	UserID      string `json:"-"`
	CustomerID  string `json:"customerId"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	PhoneNumber string `json:"phoneNumber"`
	Code        string `json:"code"`
	From        string `json:"from"`
	To          string `json:"to"`
	InternetFee string `json:"internetFee"`
	Tax         string `json:"tax"`
	Total       string `json:"total"`
}

// GetAllProviders returns all bill providers.
func (p *Provider) GetAllProviders(ctx context.Context) ([]ServiceProvider, error) {
	rows, err := p.dbSql.GetPmConnection().QueryContext(ctx, "SELECT id, name FROM provider ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var providers []ServiceProvider
	for rows.Next() {
		var sp ServiceProvider
		if err := rows.Scan(&sp.ID, &sp.Name); err != nil {
			return nil, err
		}
		providers = append(providers, sp)
	}
	return providers, rows.Err()
}

// GetInternetBillByUserID returns the internet bill for a specific user.
func (p *Provider) GetInternetBillByUserID(ctx context.Context, userID string) (*InternetBill, error) {
	row := p.dbSql.GetPmConnection().QueryRowContext(ctx,
		"SELECT id, user_id, customer_id, name, address, phone_number, code, bill_from, bill_to, internet_fee, tax, total FROM internet_bill WHERE user_id = $1",
		userID,
	)

	var bill InternetBill
	err := row.Scan(
		&bill.ID, &bill.UserID, &bill.CustomerID, &bill.Name,
		&bill.Address, &bill.PhoneNumber, &bill.Code,
		&bill.From, &bill.To, &bill.InternetFee, &bill.Tax, &bill.Total,
	)
	if err == sql.ErrNoRows {
		return nil, NotFound("internet_bill", userID)
	}
	if err != nil {
		return nil, err
	}
	return &bill, nil
}
