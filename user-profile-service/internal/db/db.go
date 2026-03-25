package db

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

// NewDB creates a new database connection from DATABASE_URL.
// Pattern from: addons-issuance-lc-service/server/core_db.go
func NewDB(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}
