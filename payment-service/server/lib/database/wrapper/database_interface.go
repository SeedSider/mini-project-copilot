package databasewrapper

import "database/sql"

type DatabaseInterface interface {
	Open(driverName, dataSourceName string) (*sql.DB, error)
}

type DatabaseConnectionInterface interface {
	Ping() error
	Close() error
	SetMaxIdleConns(n int)
	SetMaxOpenConns(n int)
	Begin() (*sql.Tx, error)
}
