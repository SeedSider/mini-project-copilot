package db

import (
	"github.com/bankease/payment-service/server/lib/database"
	"github.com/bankease/payment-service/server/lib/logger"
)

var log *logger.Logger

type Provider struct {
	dbSql *database.DbSql
}

func New(dbSql *database.DbSql, logger *logger.Logger) *Provider {
	log = logger
	return &Provider{
		dbSql: dbSql,
	}
}

func (p *Provider) GetDbSql() *database.DbSql {
	return p.dbSql
}
