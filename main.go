package main

import (
	"chirpy/handlers"
	"chirpy/internal/database"
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

	mux := http.ServeMux{}

	apiHandlers := handlers.NewAPIHandler(dbQueries)

	mux.HandleFunc("GET /api/healthz", apiHandlers.HealthCheck)
	mux.HandleFunc("POST /api/validate_chirp", apiHandlers.ValidateChirp)
	mux.HandleFunc("POST /api/users", apiHandlers.CreateUser)

	apiCfg := &APIConfig{}
	mux.Handle("/app/", apiCfg.MiddlewareMetricsInc(updateHeader(http.StripPrefix("/app", http.FileServer(http.Dir("./app"))))))

	fs := http.FileServer(http.Dir("./app/assets/"))
	mux.Handle("/app/assets", http.StripPrefix("/app/assets", fs))

	// get metrics
	mux.HandleFunc("GET /admin/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		hits := apiCfg.GetMetrics()
		_, err := w.Write([]byte(fmt.Sprintf(
			`
        <html>
          <body>
            <h1>Welcome, Chirpy Admin</h1>
            <p>Chirpy has been visited %s times!</p>
          </body>
        </html>
      `, hits,
		)))
		if err != nil {
			log.Fatal(err)
		}
	})

	mux.HandleFunc("POST /admin/reset", func(w http.ResponseWriter, r *http.Request) {
		if os.Getenv("PLATFORM") != "dev" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		err := dbQueries.DeleteUsers(r.Context())
		if err != nil {
			log.Println(err)
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_ = apiCfg.ResetMetrics()
		_, err = w.Write([]byte("OK"))
		if err != nil {
			log.Fatal(err)
		}
	})

	srv := &http.Server{Addr: ":8080", Handler: &mux}

	err = srv.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
