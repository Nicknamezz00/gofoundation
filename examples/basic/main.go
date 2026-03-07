package main

import (
	"log"
	"net/http"

	"github.com/Nicknamezz00/gofoundation/gateway"
	"github.com/Nicknamezz00/gofoundation/response"
)

func main() {
	// Create gateway with minimal configuration
	gw, err := gateway.New(gateway.DefaultConfig("basic-example"))
	if err != nil {
		log.Fatal(err)
	}

	// Define handlers
	http.HandleFunc("/api/users", gw.HandlerFunc(handleUsers))
	http.HandleFunc("/api/health", gw.HandlerFunc(handleHealth))

	// Start server
	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	rw := w.(response.Writer)

	users := []map[string]interface{}{
		{"id": 1, "name": "Alice"},
		{"id": 2, "name": "Bob"},
	}

	rw.WriteJSON(http.StatusOK, users)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	rw := w.(response.Writer)

	health := map[string]interface{}{
		"status": "healthy",
	}

	rw.WriteJSON(http.StatusOK, health)
}
