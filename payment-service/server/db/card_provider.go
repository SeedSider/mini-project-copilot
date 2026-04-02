package db

import (
	"context"

	"github.com/lib/pq"
)

// PaymentCard represents a card/account for prepaid payments.
type PaymentCard struct {
	ID             string   `json:"id"`
	AccountID      string   `json:"accountId"`
	HolderName     string   `json:"holderName"`
	CardLabel      string   `json:"cardLabel"`
	MaskedNumber   string   `json:"maskedNumber"`
	Balance        int64    `json:"balance"`
	Currency       string   `json:"currency"`
	Brand          string   `json:"brand"`
	GradientColors []string `json:"gradientColors"`
}

// GetCardsByAccountID returns all payment cards for a given account.
func (p *Provider) GetCardsByAccountID(ctx context.Context, accountID string) ([]PaymentCard, error) {
	rows, err := p.dbSql.GetPmConnection().QueryContext(ctx,
		`SELECT id, account_id, holder_name, card_label, masked_number, balance, currency, brand, gradient_colors
		 FROM payment_card WHERE account_id = $1`,
		accountID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []PaymentCard
	for rows.Next() {
		var c PaymentCard
		if err := rows.Scan(&c.ID, &c.AccountID, &c.HolderName, &c.CardLabel, &c.MaskedNumber,
			&c.Balance, &c.Currency, &c.Brand, pq.Array(&c.GradientColors)); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return cards, rows.Err()
}

// CreatePaymentCard inserts a new payment card and returns the created record.
func (p *Provider) CreatePaymentCard(ctx context.Context, card PaymentCard) (*PaymentCard, error) {
	var created PaymentCard
	err := p.dbSql.GetPmConnection().QueryRowContext(ctx,
		`INSERT INTO payment_card (account_id, holder_name, card_label, masked_number, balance, currency, brand, gradient_colors)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, account_id, holder_name, card_label, masked_number, balance, currency, brand, gradient_colors`,
		card.AccountID, card.HolderName, card.CardLabel, card.MaskedNumber,
		card.Balance, card.Currency, card.Brand, pq.Array(card.GradientColors),
	).Scan(&created.ID, &created.AccountID, &created.HolderName, &created.CardLabel, &created.MaskedNumber,
		&created.Balance, &created.Currency, &created.Brand, pq.Array(&created.GradientColors))
	if err != nil {
		return nil, err
	}
	return &created, nil
}
