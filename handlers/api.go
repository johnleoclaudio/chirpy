package handlers

import (
	"chirpy/internal/auth"
	"chirpy/internal/config"
	"chirpy/internal/database"
	"chirpy/utils"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"database/sql"

	"github.com/google/uuid"
)

type Chirp struct {
	Body string `json:"body"`
}

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginParams struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type LoginResponse struct {
	database.User
	Token        string `json:"token"`
	RefreshToken string `json:"refresh_token"`
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

func (a *APIHandlerStruct) CreateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var user User
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&user)
	if err != nil {
		log.Println("failed to decode data", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	hashedPassword, err := auth.HashPassword(user.Password)
	if err != nil {
		log.Printf("failed to hash user password: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	createUserParam := &database.CreateUserParams{
		Email:          user.Email,
		HashedPassword: hashedPassword,
	}

	createdUser, err := a.DBQueries.CreateUser(r.Context(), *createUserParam)
	if err != nil {
		log.Println("failed to create user", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonData, _ := json.Marshal(createdUser)

	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write([]byte(jsonData))
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
