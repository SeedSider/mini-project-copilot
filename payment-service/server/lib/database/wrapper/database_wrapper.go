package databasewrapper

import "database/sql"

type DatabaseWrapper struct {
	DatabaseInterface
}

func (dw *DatabaseWrapper) Open(driverName, dataSourceName string) (*sql.DB, error) {
	return sql.Open(driverName, dataSourceName)
}
