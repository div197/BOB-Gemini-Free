package main

import (
	"net/http"
	"time"
)

// newGatewayHTTPServer builds the listener for the gateway. Read limits stop a
// client from stalling a request body or holding an idle keep-alive forever.
// WriteTimeout stays zero: SSE streams are long-lived and a write deadline would
// cut them mid-answer.
func newGatewayHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       5 * time.Minute,
		IdleTimeout:       2 * time.Minute,
		WriteTimeout:      0,
	}
}
