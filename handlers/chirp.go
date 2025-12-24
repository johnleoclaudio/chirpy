package handlers

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"chirpy/utils"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
)

func (a *APIHandlerStruct) CreateChirp(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("GetBearerToken error: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userUUID, err := auth.ValidateJWT(token, a.APIConfig.JWTSecret)
	if err != nil {
		log.Printf("failed to validate JWT: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var chirpStr Chirp
	decode := json.NewDecoder(r.Body)
	err = decode.Decode(&chirpStr)
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

	userID, err := uuid.Parse(userUUID)
	if err != nil {
		log.Printf("failed to parse user UUID: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	params := database.CreateChirpParams{
		Body:   chirpStr.Body,
		UserID: userID,
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
	var chirps []database.Chirp
	var err error

	authorID := r.URL.Query().Get("author_id")

	if authorID != "" {
		userID, parseErr := uuid.Parse(authorID)
		if parseErr == nil {
			chirps, err = a.DBQueries.ListChirpsByAuthorID(r.Context(), userID)
		}
	} else {
		chirps, err = a.DBQueries.ListChirps(r.Context())
	}

	if err != nil {
		log.Printf("failed to list chrip: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		utils.RespondError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	w.WriteHeader(http.StatusOK)
	utils.RespondJSON(w, http.StatusOK, chirps)
}

func (a *APIHandlerStruct) GetChirp(w http.ResponseWriter, r *http.Request) {
	chirpID := r.PathValue("chirpID")
	u, err := uuid.Parse(chirpID)
	if err != nil {
		log.Fatalf("failed to parse UUID: %v", err)
	}
	chirp, err := a.DBQueries.GetChirp(r.Context(), u)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		log.Printf("failed to get chrip: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		utils.RespondError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	w.WriteHeader(http.StatusOK)
	utils.RespondJSON(w, http.StatusOK, chirp)
}

func (a *APIHandlerStruct) DeleteChirp(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		log.Printf("GetBearerToken error: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userUUID, err := auth.ValidateJWT(token, a.APIConfig.JWTSecret)
	if err != nil {
		log.Printf("failed to validate JWT: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	userID, err := uuid.Parse(userUUID)
	if err != nil {
		log.Fatalf("failed to parse user UUID: %v", err)
	}

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		log.Fatalf("failed to parse chirp UUID: %v", err)
	}

	chirp, err := a.DBQueries.GetChirp(r.Context(), chirpID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		log.Printf("failed to get chrip: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		utils.RespondError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	if chirp.UserID != userID {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	err = a.DBQueries.DeleteChirp(r.Context(), database.DeleteChirpParams{ID: chirpID, UserID: userID})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		log.Printf("failed to get chrip: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		utils.RespondError(w, http.StatusInternalServerError, "Something went wrong")
		return
	}

	w.WriteHeader(http.StatusNoContent)
	utils.RespondJSON(w, http.StatusOK, chirp)
}
