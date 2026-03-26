package db

import (
	"database/sql"
	"embed"
	"log"
	"sort"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

// RunMigrations auto-runs all SQL migration files in order on startup.
// Each migration must be idempotent (IF NOT EXISTS / IF NOT EXISTS).
func RunMigrations(db *sql.DB) error {
	entries, err := migrationFS.ReadDir("migrations")
	if err != nil {
		return err
	}

	// Ensure files are processed in lexicographic (numeric) order.
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		content, err := migrationFS.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return err
		}
		if _, err = db.Exec(string(content)); err != nil {
			return err
		}
		log.Printf("Migration applied: %s", entry.Name())
	}

	log.Println("All database migrations applied successfully")
	return nil
}
