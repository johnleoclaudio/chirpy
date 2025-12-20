package main

import (
	"chirpy/handlers"
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

	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	dbQueries := database.New(db)
	apiMetrics := metrics.NewAPIMetrics()

	mux := http.ServeMux{}

	apiMiddlewares := middlewares.NewMiddlwares(apiMetrics)
	apiHandlers := handlers.NewAPIHandler(dbQueries)
	adminHandlers := handlers.NewAdminHandlers(os.Getenv("PLATFORM"), apiMetrics, dbQueries)

	mux.HandleFunc("GET /api/healthz", apiHandlers.HealthCheck)
	mux.HandleFunc("GET /api/chirps", apiHandlers.ListChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiHandlers.GetChirp)
	mux.HandleFunc("POST /api/chirps", apiHandlers.CreateChirp)
	mux.HandleFunc("POST /api/users", apiHandlers.CreateUser)

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
