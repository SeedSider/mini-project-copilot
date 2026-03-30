package server

import (
	"net/http"

	"github.com/bankease/user-profile-service/internal/handlers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"
)

// setupRoutes configures all routes and middleware.
// Pattern from: addons-issuance-lc-service/server/gateway_http_handler.go
func setupRoutes(profileHandler *handlers.ProfileHandler, menuHandler *handlers.MenuHandler, uploadHandler *handlers.UploadHandler, searchHandler *handlers.SearchHandler) chi.Router {
	r := chi.NewRouter()

	// Middleware stack
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	// Profile routes
	r.Get("/api/profile", profileHandler.GetMyProfile)
	r.Post("/api/profile", profileHandler.CreateProfile)
	r.Get("/api/profile/user/{user_id}", profileHandler.GetProfileByUserID)
	r.Get("/api/profile/{id}", profileHandler.GetProfile)
	r.Put("/api/profile/{id}", profileHandler.UpdateProfile)

	// Menu routes
	r.Get("/api/menu", menuHandler.GetAllMenus)
	r.Get("/api/menu/{accountType}", menuHandler.GetMenusByAccountType)

	// Upload routes
	r.Post("/api/upload/image", uploadHandler.UploadImage)

	// Search / rates / branches routes
	r.Get("/api/exchange-rates", searchHandler.GetExchangeRates)
	r.Get("/api/interest-rates", searchHandler.GetInterestRates)
	r.Get("/api/branches", searchHandler.GetBranches)

	// Swagger UI — host is omitted from the spec so Swagger UI automatically
	// uses whatever host the browser is currently on (localhost, VM IP, etc.)
	r.Get("/swagger/*", httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
	))

	return r
}

// corsMiddleware sets CORS headers for frontend (Expo web) compatibility.
// Pattern from: addons-issuance-lc-service/server/gateway_http_handler.go cors()
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Idempotency-Key")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
