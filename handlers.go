package main

import (
	"fmt"
	"net/http"
)

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	// Get the current value
	count := cfg.fileserverHits.Load()

	w.Header().Set("Content-Type", "text/html")

	// Write it to the response
	// Format HTML with the metrics count
	html := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", count)

	// Write the HTML to the response
	fmt.Fprint(w, html)
}

func (cfg *apiConfig) middlewareMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) resetMetricsHandler(w http.ResponseWriter, r *http.Request) {
	cfg.fileserverHits.Store(0)

	if cfg.env == "dev" {
		cfg.DB.DeleteUsers(r.Context())
		err := cfg.DB.DeleteUsers(r.Context())
		if err != nil {
			http.Error(w, "Couldn't delete users", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusForbidden)
	}
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
