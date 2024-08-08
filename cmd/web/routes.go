package main

import (
	"net/http"

	"github.com/bmizerany/pat"
	"github.com/resonantchaos22/go-ws/internal/handlers"
)

// this function returns a Handler which can be served at a port using `ListenAndServe`
func routes() http.Handler {
	mux := pat.New()

	mux.Get("/", http.HandlerFunc(handlers.Home))
	mux.Get("/ws", http.HandlerFunc(handlers.WsEndpoint))

	return mux
}
