package handlers

import (
	"chirpy/internal/database"
	"chirpy/metrics"
	"fmt"
	"log"
	"net/http"
)

type AdminHandlerStruct struct {
	Env        string
	APIMetrics *metrics.API
	DBQueries  *database.Queries
}

func NewAdminHandlers(env string, apiMetrics *metrics.API, dbQueries *database.Queries) *AdminHandlerStruct {
	return &AdminHandlerStruct{
		Env:        env,
		APIMetrics: apiMetrics,
		DBQueries:  dbQueries,
	}
}

func (a *AdminHandlerStruct) GetMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	hits := a.APIMetrics.GetMetrics()
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
}

func (a *AdminHandlerStruct) Reset(w http.ResponseWriter, r *http.Request) {
	if a.Env != "dev" {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	err := a.DBQueries.DeleteUsers(r.Context())
	if err != nil {
		log.Println(err)
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_ = a.APIMetrics.ResetMetrics()
	_, err = w.Write([]byte("OK"))
	if err != nil {
		log.Fatal(err)
	}
}
