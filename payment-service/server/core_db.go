package main

import (
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/bankease/payment-service/migrations"
	"github.com/bankease/payment-service/server/lib/database"
	databasewrapper "github.com/bankease/payment-service/server/lib/database/wrapper"
)

var (
	dbSql *database.DbSql
)

func startDBConnection() {
	log.Info("", "startDBConnection", "Starting Db Connections", nil, nil, nil, nil)
	initDBMain()
}

func closeDBConnection() {
	closeDBMain()
}

func initDBMain() {
	log.Info("", "initDBMain", "Main Db - Connecting", nil, nil, nil, nil)

	maxRetry, convErr := strconv.Atoi(config.DbMaxRetry)
	if convErr != nil {
		maxRetry = 3
	}

	dbTimeout, convErr := strconv.Atoi(config.DbTimeout)
	if convErr != nil {
		dbTimeout = 120
		log.Info("", "initDBMain", fmt.Sprintf("Failed to convert database Timeout, set to default: %ds", dbTimeout), nil, nil, nil, convErr)
	}

	dbSql = database.InitConnectionDB("postgres", database.Config{
		Host:         config.DbHost,
		Port:         config.DbPort,
		User:         config.DbUser,
		Password:     config.DbPassword,
		DatabaseName: config.DbName,
		SslMode:      config.DbSslmode,
		TimeZone:     config.DbTimezone,
		MaxRetry:     maxRetry,
		Timeout:      time.Duration(dbTimeout) * time.Second,
	}, &databasewrapper.DatabaseWrapper{})

	err := dbSql.Connect()
	if err != nil {
		log.Fatal("", "initDBMain", fmt.Sprintf("[initDBMain] Failed connect to DB main: %v", err), nil, nil, nil, err)
		os.Exit(1)
		return
	}

	dbSql.SetMaxIdleConns(0)
	dbSql.SetMaxOpenConns(100)

	log.Info("", "initDBMain", "Main Db - Connected", nil, nil, nil, nil)
}

func closeDBMain() {
	if err := dbSql.ClosePmConnection(); err != nil {
		log.Error("", "closeDBMain", fmt.Sprintf("Error on disconnection with DB Main : %v", err), nil, nil, nil, err)
	}
	log.Info("", "closeDBMain", "Closing DB Main Success", nil, nil, nil, nil)
}

func runMigration() {
	log.Info("", "runMigration", "Running database migration...", nil, nil, nil, nil)

	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		log.Fatal("", "runMigration", fmt.Sprintf("Failed to read migrations: %v", err), nil, nil, nil, err)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		content, err := fs.ReadFile(migrations.FS, entry.Name())
		if err != nil {
			log.Fatal("", "runMigration", fmt.Sprintf("Failed to read migration %s: %v", entry.Name(), err), nil, nil, nil, err)
		}

		_, err = dbSql.GetPmConnection().Exec(string(content))
		if err != nil {
			log.Fatal("", "runMigration", fmt.Sprintf("Migration %s failed: %v", entry.Name(), err), nil, nil, nil, err)
		}

		log.Info("", "runMigration", fmt.Sprintf("Applied migration: %s", entry.Name()), nil, nil, nil, nil)
	}

	log.Info("", "runMigration", "Migration completed", nil, nil, nil, nil)
}
