package main

import (
	"chirpy/handlers"
	"chirpy/internal/database"
	"chirpy/metrics"
	"chirpy/middlewares"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type APIConfig struct {
	FileserverHits atomic.Int32
}

func (c *APIConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.FileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (c *APIConfig) GetMetrics() string {
	hits := c.FileserverHits.Load()
	return fmt.Sprintf("%v", hits)
}

func (c *APIConfig) ResetMetrics() bool {
	hits := c.FileserverHits.Load()
	success := c.FileserverHits.CompareAndSwap(hits, 0)
	return success
}

func updateHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache")
		next.ServeHTTP(w, r)
	})
}
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
	mux.HandleFunc("POST /api/chirps", apiHandlers.CreateChirp)
	mux.HandleFunc("POST /api/users", apiHandlers.CreateUser)

	mux.Handle("/app/", apiMiddlewares.MiddlewareMetricsInc(updateHeader(http.StripPrefix("/app", http.FileServer(http.Dir("./app"))))))

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
