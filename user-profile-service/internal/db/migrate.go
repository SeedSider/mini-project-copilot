package db

import (
	"database/sql"
	"embed"
	"log"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// RunMigrations auto-runs SQL migrations on startup.
// Idempotent via CREATE TABLE IF NOT EXISTS.
func RunMigrations(db *sql.DB) error {
	content, err := migrationFS.ReadFile("migrations/001_init.sql")
	if err != nil {
		return err
	}

	_, err = db.Exec(string(content))
	if err != nil {
		return err
	}

	log.Println("Database migrations applied successfully")
	return nil
}
