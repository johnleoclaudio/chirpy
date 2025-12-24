package handlers

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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

func (a *APIHandlerStruct) UpdateUser(w http.ResponseWriter, r *http.Request) {
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

	var user User
	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&user)
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

	userID, err := uuid.Parse(userUUID)
	if err != nil {
		log.Printf("failed to parse user UUID: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	updatedUser, err := a.DBQueries.UpdateUser(r.Context(), database.UpdateUserParams{
		ID:             userID,
		Email:          user.Email,
		HashedPassword: hashedPassword,
	})

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			log.Printf("user ID not found: %v", err)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonData, _ := json.Marshal(updatedUser)

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(jsonData))
}
