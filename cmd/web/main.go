package main

import (
	"log"
	"net/http"

	"github.com/resonantchaos22/go-ws/internal/handlers"
)

func main() {
	mux := routes()

	log.Println("Starting Channel Listener...")
	go handlers.ListenToWsChannel()

	log.Println("Starting server on port 8080...")

	http.ListenAndServe(":8080", mux)
}
