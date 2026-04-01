package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// PrepaidTransaction represents a mobile prepaid top-up transaction.
type PrepaidTransaction struct {
	ID             string `json:"id"`
	CardID         string `json:"-"`
	Phone          string `json:"-"`
	Amount         int64  `json:"-"`
	Status         string `json:"status"`
	Message        string `json:"message"`
	IdempotencyKey string `json:"-"`
	Timestamp      string `json:"timestamp"`
}

// GetTransactionByIdempotencyKey looks up an existing transaction by its idempotency key.
func (p *Provider) GetTransactionByIdempotencyKey(ctx context.Context, key string) (*PrepaidTransaction, error) {
	row := p.dbSql.GetPmConnection().QueryRowContext(ctx,
		"SELECT id, status, message, created_at FROM prepaid_transaction WHERE idempotency_key = $1",
		key,
	)

	var txn PrepaidTransaction
	var createdAt time.Time
	err := row.Scan(&txn.ID, &txn.Status, &txn.Message, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	txn.Timestamp = createdAt.UTC().Format(time.RFC3339Nano)
	return &txn, nil
}

// CreatePrepaidTransaction inserts a new prepaid transaction.
func (p *Provider) CreatePrepaidTransaction(ctx context.Context, txn PrepaidTransaction) (*PrepaidTransaction, error) {
	now := time.Now().UTC()
	_, err := p.dbSql.GetPmConnection().ExecContext(ctx,
		`INSERT INTO prepaid_transaction (id, card_id, phone, amount, status, message, idempotency_key, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		txn.ID, txn.CardID, txn.Phone, txn.Amount, txn.Status, txn.Message, txn.IdempotencyKey, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert prepaid transaction: %w", err)
	}

	txn.Timestamp = now.Format(time.RFC3339Nano)
	return &txn, nil
}
