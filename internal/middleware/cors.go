package middleware

import (
	"github.com/go-chi/cors"
)

// CORS returns cors.Options parameterized by the given allowed origins.
// If "*" is present, AllowCredentials is set to false (browsers reject
// Access-Control-Allow-Credentials: true with a wildcard origin).
func CORS(allowedOrigins []string) cors.Options {
	if len(allowedOrigins) == 0 {
		allowedOrigins = []string{"http://localhost:3000"}
	}

	allowCreds := true
	for _, o := range allowedOrigins {
		if o == "*" {
			allowCreds = false
			break
		}
	}

	return cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: allowCreds,
		MaxAge:           300,
	}
}
