package handlers

import (
	"chirpy/internal/database"
	"chirpy/utils"
	"encoding/json"
	"log"
	"net/http"
)

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

func (a *APIHandlerStruct) ValidateChirp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	type chirp struct {
		Body string `json:"body"`
	}

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
		log.Printf("failed to decode data: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		utils.RespondError(w, http.StatusInternalServerError, &errorResp{Error: "Something went wrong"})
		return
	}

	if len(chirpStr.Body) > 140 {
		w.WriteHeader(http.StatusBadRequest)
		utils.RespondError(w, http.StatusInternalServerError, &errorResp{Error: "Chirp is too long"})
		return
	}

	utils.RespondJSON(w, http.StatusOK, &successResp{CleanedBody: utils.ProfaneFilter(chirpStr.Body)})
}
