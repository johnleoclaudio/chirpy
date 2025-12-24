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
	"time"
)

type LoginParams struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type LoginResponse struct {
	database.User
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
}

func (a *APIHandlerStruct) Login(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var loginParams LoginParams
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&loginParams)
	if err != nil {
		log.Println("failed to decode data", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	retrievedUser, err := a.DBQueries.GetUser(r.Context(), loginParams.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		log.Println("failed to get user", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	authenticated, err := auth.CheckPassword(loginParams.Password, retrievedUser.HashedPassword)
	if err != nil {
		log.Println("failed to check password", loginParams.Password, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !authenticated {
		utils.RespondError(w, http.StatusUnauthorized, "Incorrect email or password")
		return
	}

	token, err := auth.MakeRefreshToken()
	if err != nil {
		log.Println("failed to create refreshToken", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	refreshToken, err := a.DBQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{Token: token, UserID: retrievedUser.ID})
	if err != nil {
		log.Println("failed to insert refresh token", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	expirationInSeconds := 3600
	jwtToken, err := auth.MakeJWT(retrievedUser.ID, a.APIConfig.JWTSecret, time.Second*time.Duration(expirationInSeconds))
	if err != nil {
		log.Println("failed to create JWT", err)
		utils.RespondError(w, http.StatusInternalServerError, "failed to create JWT")
	}

	log.Printf("created access token: %s, owner: %s", jwtToken, refreshToken.UserID)

	retrievedUser.HashedPassword = ""

	utils.RespondJSON(w, http.StatusOK, &LoginResponse{
		User:         retrievedUser,
		Token:        jwtToken,
		RefreshToken: refreshToken.Token,
	})
}

func (a *APIHandlerStruct) RefreshAccessToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetRefreshTokenHeader(r.Header)
	if err != nil {
		log.Printf("failed to get refresh token header: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token, err := a.DBQueries.GetRefreshToken(r.Context(), refreshToken)
	if err != nil {
		log.Printf("failed to get refresh token from db: %v, %s", err, refreshToken)
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// check if revoked_at is not empty or
	// if the current time and date is after the token expiration
	if !token.RevokedAt.Time.IsZero() || time.Now().After(token.ExpiresAt) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	authToken, err := auth.MakeJWT(token.UserID, a.APIConfig.JWTSecret, time.Second*3600)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var payload struct {
		Token string `json:"token"`
	}
	payload.Token = authToken
	utils.RespondJSON(w, http.StatusOK, payload)
}

func (a *APIHandlerStruct) RevokeRefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetRefreshTokenHeader(r.Header)
	if err != nil {
		log.Printf("failed to get refresh token header: %v", err)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	_, err = a.DBQueries.RevokeRefreskToken(r.Context(), refreshToken)
	if err != nil {
		log.Printf("failed to revoke refresh token from db: %v, %s", err, refreshToken)
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
