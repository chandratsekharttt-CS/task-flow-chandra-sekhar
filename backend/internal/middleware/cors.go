package middleware

import (
	"github.com/go-chi/cors"
)

// CORSOptions returns the CORS configuration for the API.
func CORSOptions() cors.Options {
	return cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173", "http://web:3000"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}
}
