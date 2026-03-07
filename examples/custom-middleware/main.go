package main

import (
	"log"
	"net/http"

	"github.com/Nicknamezz00/gofoundation/gateway"
	"github.com/Nicknamezz00/gofoundation/response"
)

func main() {
	// Create gateway
	config := gateway.DefaultConfig("custom-middleware-example")

	// Add custom middleware
	config.Middlewares = []gateway.Middleware{
		authMiddleware,
	}

	gw, err := gateway.New(config)
	if err != nil {
		log.Fatal(err)
	}

	// Define handlers
	http.HandleFunc("/api/protected", gw.HandlerFunc(handleProtected))
	http.HandleFunc("/api/public", gw.HandlerFunc(handlePublic))

	// Start server
	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

// authMiddleware checks for authorization header
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth for public endpoints
		if r.URL.Path == "/api/public" {
			next.ServeHTTP(w, r)
			return
		}

		token := r.Header.Get("Authorization")
		if token == "" {
			rw := w.(response.Writer)
			rw.WriteError(http.StatusUnauthorized, http.ErrNotSupported)
			return
		}

		// In real app, validate token here

		next.ServeHTTP(w, r)
	})
}

func handleProtected(w http.ResponseWriter, r *http.Request) {
	rw := w.(response.Writer)
	rw.WriteJSON(http.StatusOK, map[string]string{
		"message": "This is protected data",
	})
}

func handlePublic(w http.ResponseWriter, r *http.Request) {
	rw := w.(response.Writer)
	rw.WriteJSON(http.StatusOK, map[string]string{
		"message": "This is public data",
	})
}
