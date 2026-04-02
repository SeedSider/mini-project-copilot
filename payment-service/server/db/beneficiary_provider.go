package db

import (
	"context"
	"fmt"
)

// Beneficiary represents a saved mobile top-up contact.
type Beneficiary struct {
	ID        string `json:"id"`
	AccountID string `json:"-"`
	Name      string `json:"name"`
	Phone     string `json:"phone"`
	Avatar    string `json:"avatar"`
}

// GetBeneficiariesByAccountID returns all beneficiaries for a given account.
func (p *Provider) GetBeneficiariesByAccountID(ctx context.Context, accountID string) ([]Beneficiary, error) {
	rows, err := p.dbSql.GetPmConnection().QueryContext(ctx,
		"SELECT id, account_id, name, phone, avatar FROM beneficiary WHERE account_id = $1",
		accountID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var beneficiaries []Beneficiary
	for rows.Next() {
		var b Beneficiary
		if err := rows.Scan(&b.ID, &b.AccountID, &b.Name, &b.Phone, &b.Avatar); err != nil {
			return nil, err
		}
		beneficiaries = append(beneficiaries, b)
	}
	return beneficiaries, rows.Err()
}

// CreateBeneficiary inserts a new beneficiary and returns the created record.
func (p *Provider) CreateBeneficiary(ctx context.Context, b Beneficiary) (*Beneficiary, error) {
	var created Beneficiary
	err := p.dbSql.GetPmConnection().QueryRowContext(ctx,
		`INSERT INTO beneficiary (account_id, name, phone, avatar)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, account_id, name, phone, avatar`,
		b.AccountID, b.Name, b.Phone, b.Avatar,
	).Scan(&created.ID, &created.AccountID, &created.Name, &created.Phone, &created.Avatar)
	if err != nil {
		return nil, fmt.Errorf("failed to insert beneficiary: %w", err)
	}
	return &created, nil
}

// SearchBeneficiaries searches beneficiaries by name or phone (case-insensitive partial match).
func (p *Provider) SearchBeneficiaries(ctx context.Context, accountID, query string) ([]Beneficiary, error) {
	likeQuery := "%" + query + "%"
	rows, err := p.dbSql.GetPmConnection().QueryContext(ctx,
		`SELECT id, account_id, name, phone, avatar FROM beneficiary
		 WHERE account_id = $1 AND (name ILIKE $2 OR phone ILIKE $2)`,
		accountID, likeQuery,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var beneficiaries []Beneficiary
	for rows.Next() {
		var b Beneficiary
		if err := rows.Scan(&b.ID, &b.AccountID, &b.Name, &b.Phone, &b.Avatar); err != nil {
			return nil, err
		}
		beneficiaries = append(beneficiaries, b)
	}
	return beneficiaries, rows.Err()
}
