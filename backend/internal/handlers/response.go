package handlers

import (
	"encoding/json"
	"net/http"
)

// respondJSON writes a JSON response with the given status code.
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

// respondError writes a JSON error response.
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

// respondValidationError writes a 400 response with structured field errors.
func respondValidationError(w http.ResponseWriter, fields map[string]string) {
	respondJSON(w, http.StatusBadRequest, map[string]interface{}{
		"error":  "validation failed",
		"fields": fields,
	})
}
