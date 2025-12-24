package handlers

import (
	"chirpy/internal/auth"
	"chirpy/internal/database"
	"encoding/json"
	"log"
	"net/http"
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
