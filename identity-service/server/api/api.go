package api

import (
	"time"

	"bitbucket.bri.co.id/scm/addons/addons-identity-service/server/db"
	manager "bitbucket.bri.co.id/scm/addons/addons-identity-service/server/jwt"
	"bitbucket.bri.co.id/scm/addons/addons-identity-service/server/lib/database"
	"bitbucket.bri.co.id/scm/addons/addons-identity-service/server/lib/logger"
)

var log *logger.Logger

type Server struct {
	provider *db.Provider
	manager  *manager.JWTManager
	logger   *logger.Logger
}

func New(
	jwtSecret string,
	jwtDuration string,
	dbConnection *database.DbSql,
	logger *logger.Logger,
	provider *db.Provider,
) *Server {
	log = logger
	tokenDuration, err := time.ParseDuration(jwtDuration)
	if err != nil {
		log.Fatal("", "New", "Failed when Parsing Duration", nil, nil, nil, err)
	}

	return &Server{
		provider: db.New(dbConnection, logger),
		manager:  manager.NewJWTManager(jwtSecret, tokenDuration),
		logger:   logger,
	}
}

func (s *Server) GetManager() *manager.JWTManager {
	return s.manager
}
