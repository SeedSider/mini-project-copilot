package main

import (
	"database/sql"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/bankease/user-profile-service/migrations"
	_ "github.com/lib/pq"
)

var dbSql *sql.DB

func startDBConnection() {
	if config.DatabaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	db, err := sql.Open("postgres", config.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	dbSql = db
	log.Println("Database connected")
}

func closeDBConnection() {
	if dbSql != nil {
		dbSql.Close()
		log.Println("Database connection closed")
	}
}

func runMigration() {
	entries, err := migrations.FS.ReadDir(".")
	if err != nil {
		log.Fatalf("Failed to read migrations: %v", err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		content, err := migrations.FS.ReadFile(entry.Name())
		if err != nil {
			log.Fatalf("Failed to read migration %s: %v", entry.Name(), err)
		}
		if _, err = dbSql.Exec(string(content)); err != nil {
			log.Fatalf("Migration %s failed: %v", entry.Name(), err)
		}
		log.Printf("Migration applied: %s", entry.Name())
	}

	log.Println("All database migrations applied successfully")
}
