package main

import (
	"io"
	"log"
	"net/http"
	"time"
)

func handleRequest(w http.ResponseWriter, r *http.Request) {
	targetURL := r.URL
	proxyReq, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL.String(), r.Body)
	if err != nil {
		http.Error(w, "Error creating proxy request", http.StatusInternalServerError)
		return
	}

	for name, values := range r.Header {
		for _, value := range values {
			proxyReq.Header.Add(name, value)
		}
	}

	tp := http.DefaultTransport

	resp, err := tp.RoundTrip(proxyReq)
	if err != nil {
		http.Error(w, "Error sending proxy request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	w.WriteHeader(resp.StatusCode)

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Printf("failed to copy: %v", err)
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	const idleTimeout = 30 * time.Second
	const writeTimeout = 1 * time.Second
	const readTimeout = 1 * time.Second
	const readHeaderTimeout = 2 * time.Second
	server := http.Server{
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		ReadHeaderTimeout: readHeaderTimeout,

		Addr:    ":8787",
		Handler: http.HandlerFunc(handleRequest),
	}

	err := server.ListenAndServe()
	if err != nil {
		panic(err)
	}
}
