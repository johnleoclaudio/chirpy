package main

import (
	"chirpy/internal/database"
	"chirpy/utils"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	FileserverHits atomic.Int32
}

func (c *apiConfig) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.FileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (c *apiConfig) GetMetrics() string {
	hits := c.FileserverHits.Load()
	return fmt.Sprintf("%v", hits)
}

func (c *apiConfig) ResetMetrics() bool {
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
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	log.Println("DB_URL: ", dbURL)
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}
	dbQueries := database.New(db)

	mux := http.ServeMux{}

	// readiness endpoint
	// returns 200 OK if the server is ready to accept requests
	mux.HandleFunc("GET /api/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte("OK"))
		if err != nil {
			log.Fatal(err)
		}
	})

	mux.HandleFunc("POST /api/validate_chirp", func(w http.ResponseWriter, r *http.Request) {
		type chirp struct {
			Body string `json:"body"`
		}
		defer r.Body.Close()

		type errorResp struct {
			Error string `json:"error"`
		}

		type successResp struct {
			CleanedBody string `json:"cleaned_body"`
		}

		var chirpStr chirp
		decode := json.NewDecoder(r.Body)
		err := decode.Decode(&chirpStr)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			errBody := &errorResp{Error: "Something went wrong"}
			data, err := json.Marshal(errBody)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}
			w.Write([]byte(data))
			return
		}

		if len(chirpStr.Body) > 140 {
			w.WriteHeader(http.StatusBadRequest)
			errBody := &errorResp{Error: "Chirp is too long"}
			data, err := json.Marshal(errBody)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Write([]byte(data))
			return
		}
		data, err := json.Marshal(&successResp{CleanedBody: utils.ProfaneFilter(chirpStr.Body)})

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(data))
	})

	mux.HandleFunc("POST /api/users", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		type user struct {
			Email string `json:"email"`
		}

		var userBody user
		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&userBody)
		if err != nil {
			log.Println("failed to decode data", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		createdUser, err := dbQueries.CreateUser(r.Context(), userBody.Email)
		if err != nil {
			log.Println("failed to create user", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		jsonData, _ := json.Marshal(createdUser)

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(jsonData))
	})

	apiCfg := &apiConfig{}
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
