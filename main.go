package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/ValentinoFilipetto/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	DB             *database.Queries
	env            string
}

func main() {
	// Postgres connection
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)

	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		return
	}

	dbQueries := database.New(db)

	// Server configuration
	mux := http.NewServeMux()
	apiCfg := apiConfig{
		DB:  dbQueries,
		env: os.Getenv("PLATFORM"),
	}

	fileServer := http.FileServer(http.Dir("."))
	mux.Handle("/app/", apiCfg.middlewareMetrics(http.StripPrefix("/app", fileServer)))
	mux.HandleFunc("GET /api/healthz", readinessHandler)
	mux.HandleFunc("GET /api/chirps", apiCfg.GetChirpsHandler)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.GetChirpHandler)
	mux.HandleFunc("POST /api/chirps", apiCfg.AddChirpHandler)
	mux.HandleFunc("POST /api/users", apiCfg.CreateUserHandler)
	mux.HandleFunc("GET /admin/metrics", apiCfg.metricsHandler)
	mux.HandleFunc("POST /admin/reset", apiCfg.resetMetricsHandler)

	server := &http.Server{
		Addr:    "127.0.0.1:8080", // forces WSL2 to use IPv4
		Handler: mux,
	}

	fmt.Println("Server starting...")
	err = server.ListenAndServe()
	if err != nil {
		fmt.Println(err)
	}
}
