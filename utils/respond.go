package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

func RespondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	// This line does the w.Write() internally
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func RespondError(w http.ResponseWriter, status int, data interface{}) {
	RespondJSON(w, status, data)
}
