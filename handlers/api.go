package handlers

import (
	"chirpy/internal/config"
	"chirpy/internal/database"
	"log"
	"net/http"
)

type Chirp struct {
	Body string `json:"body"`
}

type APIHandlerStruct struct {
	APIConfig *config.APIConfig
	DBQueries *database.Queries
}

func NewAPIHandler(apiConfig *config.APIConfig, dbQueries *database.Queries) *APIHandlerStruct {
	return &APIHandlerStruct{
		APIConfig: apiConfig,
		DBQueries: dbQueries,
	}
}

func (a *APIHandlerStruct) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, err := w.Write([]byte("OK"))
	if err != nil {
		log.Fatal(err)
	}
}
