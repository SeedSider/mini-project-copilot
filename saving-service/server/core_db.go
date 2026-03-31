package main

import (
	"database/sql"
	"sort"
	"strings"
	"time"

	"github.com/bankease/saving-service/migrations"
	_ "github.com/lib/pq"
)

var dbSql *sql.DB

func startDBConnection() {
	log.Info("", "startDBConnection", "Starting DB connection", nil, nil, nil, nil)

	dsn := getDatabaseURL()
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("", "startDBConnection", "Failed to open database", nil, nil, nil, err)
	}

	if err := db.Ping(); err != nil {
		log.Fatal("", "startDBConnection", "Failed to ping database", nil, nil, nil, err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	dbSql = db
	log.Info("", "startDBConnection", "Database connected", nil, nil, nil, nil)
}

func closeDBConnection() {
	if dbSql != nil {
		dbSql.Close()
		log.Info("", "closeDBConnection", "Database connection closed", nil, nil, nil, nil)
	}
}

func runMigration() {
	log.Info("", "runMigration", "Running database migration...", nil, nil, nil, nil)

	entries, err := migrations.FS.ReadDir(".")
	if err != nil {
		log.Fatal("", "runMigration", "Failed to read migrations", nil, nil, nil, err)
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
			log.Fatal("", "runMigration", "Failed to read migration: "+entry.Name(), nil, nil, nil, err)
		}
		if _, err = dbSql.Exec(string(content)); err != nil {
			log.Fatal("", "runMigration", "Migration failed: "+entry.Name(), nil, nil, nil, err)
		}
		log.Info("", "runMigration", "Migration applied: "+entry.Name(), nil, nil, nil, nil)
	}

	log.Info("", "runMigration", "All database migrations applied successfully", nil, nil, nil, nil)
}
