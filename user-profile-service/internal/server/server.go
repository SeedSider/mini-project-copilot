package server

import (
	"database/sql"
	"net/http"

	"github.com/bankease/user-profile-service/internal/handlers"
	"github.com/bankease/user-profile-service/internal/repository"
	"github.com/go-chi/chi/v5"
)

// Server holds all dependencies and the HTTP router.
// Pattern from: addons-issuance-lc-service/server/main.go (DI via struct)
type Server struct {
	DB     *sql.DB
	Router chi.Router
	Port   string
}

// NewServer creates a new Server with all dependencies wired up.
func NewServer(db *sql.DB, port string, azureSASURL string, azureContainer string, jwtSecret string) *Server {
	profileRepo := &repository.ProfileRepository{DB: db}
	menuRepo := &repository.MenuRepository{DB: db}

	profileHandler := &handlers.ProfileHandler{Repo: profileRepo, JWTSecret: jwtSecret}
	menuHandler := &handlers.MenuHandler{Repo: menuRepo}
	uploadHandler := &handlers.UploadHandler{
		AzureSASURL:    azureSASURL,
		AzureContainer: azureContainer,
	}

	s := &Server{
		DB:   db,
		Port: port,
	}

	s.Router = setupRoutes(profileHandler, menuHandler, uploadHandler)
	return s
}

// Start begins listening for HTTP requests.
func (s *Server) Start() error {
	return http.ListenAndServe(":"+s.Port, s.Router)
}
