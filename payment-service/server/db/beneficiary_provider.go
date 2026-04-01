package db

import "context"

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
