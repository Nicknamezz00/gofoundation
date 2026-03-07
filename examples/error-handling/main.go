package main

import (
	"log"
	"net/http"

	"github.com/Nicknamezz00/gofoundation/errors"
	"github.com/Nicknamezz00/gofoundation/gateway"
	"github.com/Nicknamezz00/gofoundation/response"
)

func main() {
	// Create gateway
	gw, err := gateway.New(gateway.DefaultConfig("error-example"))
	if err != nil {
		log.Fatal(err)
	}

	// Define handlers
	http.HandleFunc("/api/success", gw.HandlerFunc(handleSuccess))
	http.HandleFunc("/api/error", gw.HandlerFunc(handleError))
	http.HandleFunc("/api/not-found", gw.HandlerFunc(handleNotFound))

	// Start server
	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func handleSuccess(w http.ResponseWriter, r *http.Request) {
	rw := w.(response.Writer)
	rw.WriteJSON(http.StatusOK, map[string]string{"message": "success"})
}

func handleError(w http.ResponseWriter, r *http.Request) {
	rw := w.(response.Writer)
	err := errors.InternalError("something went wrong")
	rw.WriteError(errors.GetStatusCode(err), err)
}

func handleNotFound(w http.ResponseWriter, r *http.Request) {
	rw := w.(response.Writer)
	err := errors.NotFound("resource not found")
	rw.WriteError(errors.GetStatusCode(err), err)
}
