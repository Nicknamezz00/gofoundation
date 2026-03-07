package main

import (
	"log"
	"net/http"
	"time"

	"github.com/Nicknamezz00/gofoundation/gateway"
	"github.com/Nicknamezz00/gofoundation/logger"
	"github.com/Nicknamezz00/gofoundation/response"
)

func main() {
	// Create gateway with custom logger config
	config := gateway.DefaultConfig("logging-example")
	config.Logger = logger.Config{
		Level:   logger.DebugLevel,
		DevMode: true, // Output to stdout
	}

	gw, err := gateway.New(config)
	if err != nil {
		log.Fatal(err)
	}

	// Define handlers
	http.HandleFunc("/api/process", gw.HandlerFunc(handleProcess))

	// Start server
	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func handleProcess(w http.ResponseWriter, r *http.Request) {
	rw := w.(response.Writer)

	// Simulate processing
	time.Sleep(100 * time.Millisecond)

	result := map[string]interface{}{
		"status":    "completed",
		"processed": 42,
	}

	rw.WriteJSON(http.StatusOK, result)
}
