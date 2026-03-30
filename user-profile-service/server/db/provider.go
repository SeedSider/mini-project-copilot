package db

import (
	"database/sql"
)

// Provider wraps database access for all domain queries.
// Pattern from: identity-service/server/db/provider.go
type Provider struct {
	DB *sql.DB
}

// New creates a new Provider with the given database connection.
func New(db *sql.DB) *Provider {
	return &Provider{DB: db}
}
