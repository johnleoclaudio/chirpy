package handlers

import (
	"chirpy/internal/auth"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type WebhookEvent struct {
	Event string `json:"event"`
	Data  struct {
		UserID string `json:"user_id"`
	} `json:"data"`
}

func (a *APIHandlerStruct) Webhook(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	apiKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if apiKey != a.APIConfig.PolkaKey {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var webhookEvent WebhookEvent

	decoder := json.NewDecoder(r.Body)
	err = decoder.Decode(&webhookEvent)
	if err != nil {
		log.Println("failed to decode webhook event")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if webhookEvent.Event != "user.upgraded" {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	userId, err := uuid.Parse(webhookEvent.Data.UserID)
	if err != nil {
		log.Printf("failed to parse user UUID: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = a.DBQueries.GetUserByID(r.Context(), userId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	_, err = a.DBQueries.EnableUserChirpyRed(r.Context(), userId)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
