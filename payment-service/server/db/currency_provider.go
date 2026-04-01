package db

import "context"

// Currency represents a currency entry with its exchange rate.
type Currency struct {
	Code  string  `json:"code"`
	Label string  `json:"label"`
	Rate  float64 `json:"rate"`
}

// GetAllCurrencies returns all currency entries.
func (p *Provider) GetAllCurrencies(ctx context.Context) ([]Currency, error) {
	rows, err := p.dbSql.GetPmConnection().QueryContext(ctx, "SELECT code, label, rate FROM currency ORDER BY code")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var currencies []Currency
	for rows.Next() {
		var c Currency
		if err := rows.Scan(&c.Code, &c.Label, &c.Rate); err != nil {
			return nil, err
		}
		currencies = append(currencies, c)
	}
	return currencies, rows.Err()
}
