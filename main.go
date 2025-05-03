package main

import (
	"fmt"
	"net/http"
	"sync/atomic"
)

type apiConfig struct {
	fileserverHits atomic.Int32
}

func main() {
	mux := http.NewServeMux()
	apiCfg := &apiConfig{}

	fileServer := http.FileServer(http.Dir("."))
	mux.Handle("/app/", apiCfg.middlewareMetrics(http.StripPrefix("/app", fileServer)))
	mux.HandleFunc("GET /api/healthz", readinessHandler)
	mux.HandleFunc("POST /api/validate_chirp", apiCfg.ValidateChirpHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetMetricsHandler)

	server := &http.Server{
		Addr:    "127.0.0.1:8080", // forces WSL2 to use IPv4
		Handler: mux,
	}

	fmt.Println("Server starting...")
	err := server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
}
