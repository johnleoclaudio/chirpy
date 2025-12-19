package handlers

import (
	"chirpy/internal/database"
	"chirpy/utils"
	"encoding/json"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type Chirp struct {
	Body   string `json:"body"`
	UserID string `json:"user_id,omitempty"`
}

type successResp struct {
	CleanedBody string `json:"cleaned_body"`
}

type APIHandlerStruct struct {
	DBQueries *database.Queries
}

func NewAPIHandler(dbQueries *database.Queries) *APIHandlerStruct {
	return &APIHandlerStruct{
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

func (a *APIHandlerStruct) CreateUser(w http.ResponseWriter, r *http.Request) {
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

	createdUser, err := a.DBQueries.CreateUser(r.Context(), userBody.Email)
	if err != nil {
		log.Println("failed to create user", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonData, _ := json.Marshal(createdUser)

	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(jsonData))
}

func (a *APIHandlerStruct) CreateChirp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var chirpStr Chirp
	decode := json.NewDecoder(r.Body)
	err := decode.Decode(&chirpStr)
	if err != nil {
		log.Printf("failed to decode data: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		utils.RespondError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	if len(chirpStr.Body) > 140 {
		w.WriteHeader(http.StatusBadRequest)
		utils.RespondError(w, http.StatusInternalServerError, "Chirp is too long")
		return
	}

	params := database.CreateChirpParams{
		Body:   chirpStr.Body,
		UserID: uuid.MustParse(chirpStr.UserID),
	}
	chirp, err := a.DBQueries.CreateChirp(r.Context(), params)
	if err != nil {
		log.Printf("failed to create chrip: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		utils.RespondError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	w.WriteHeader(http.StatusCreated)
	utils.RespondJSON(w, http.StatusOK, chirp)
}

func (a *APIHandlerStruct) ListChirps(w http.ResponseWriter, r *http.Request) {
	chirps, err := a.DBQueries.ListChirps(r.Context())
	if err != nil {
		log.Printf("failed to list chrip: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		utils.RespondError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	w.WriteHeader(http.StatusOK)
	utils.RespondJSON(w, http.StatusOK, chirps)
}
