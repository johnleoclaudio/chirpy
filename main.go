package main

import (
	"chirpy/handlers"
	"chirpy/internal/config"
	"chirpy/internal/database"
	"chirpy/metrics"
	"chirpy/middlewares"
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	// API Config
	apiConfig := &config.APIConfig{
		JWTSecret: os.Getenv("JWT_SECRET"),
		PolkaKey:  os.Getenv("POLKA_KEY"),
	}

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	dbQueries := database.New(db)
	apiMetrics := metrics.NewAPIMetrics()

	mux := http.ServeMux{}

	apiMiddlewares := middlewares.NewMiddlewares(apiMetrics)
	apiHandlers := handlers.NewAPIHandler(apiConfig, dbQueries)
	adminHandlers := handlers.NewAdminHandlers(os.Getenv("PLATFORM"), apiMetrics, dbQueries)

	mux.HandleFunc("GET /api/healthz", apiHandlers.HealthCheck)

	// chirps
	mux.HandleFunc("GET /api/chirps", apiHandlers.ListChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiHandlers.GetChirp)
	mux.HandleFunc("POST /api/chirps", apiHandlers.CreateChirp)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiHandlers.DeleteChirp)

	// auth
	mux.HandleFunc("POST /api/login", apiHandlers.Login)
	mux.HandleFunc("POST /api/refresh", apiHandlers.RefreshAccessToken)
	mux.HandleFunc("POST /api/revoke", apiHandlers.RevokeRefreshToken)

	// users
	mux.HandleFunc("POST /api/users", apiHandlers.CreateUser)
	mux.HandleFunc("PUT /api/users", apiHandlers.UpdateUser)

	// webhook
	mux.HandleFunc("POST /api/polka/webhooks", apiHandlers.Webhook)

	mux.Handle("/app/", apiMiddlewares.MiddlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir("./app")))))

	fs := http.FileServer(http.Dir("./app/assets/"))
	mux.Handle("/app/assets", http.StripPrefix("/app/assets", fs))

	mux.HandleFunc("GET /admin/metrics", adminHandlers.GetMetrics)
	mux.HandleFunc("POST /admin/reset", adminHandlers.Reset)

	srv := &http.Server{Addr: ":8080", Handler: &mux}

	err = srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
