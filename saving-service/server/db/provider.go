package db

import (
	"database/sql"

	"github.com/bankease/saving-service/server/lib/logger"
)

var log *logger.Logger

// Provider wraps database access for all domain queries.
type Provider struct {
	DB *sql.DB
}

// New creates a new Provider with the given database connection.
func New(db *sql.DB, logger *logger.Logger) *Provider {
	log = logger
	return &Provider{DB: db}
}
