package main

import (
   "net/http"
   "fmt"
   "sync/atomic"
  )

type apiConfig struct {
    fileserverHits atomic.Int32
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
    // Get the current value
    count := cfg.fileserverHits.Load()
    
    // Format it as a string
    responseText := fmt.Sprintf("Hits: %d", count)
    
    // Write it to the response
    w.Write([]byte(responseText))                                  
}

func (cfg *apiConfig) middlewareMetrics(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        cfg.fileserverHits.Add(1)
        next.ServeHTTP(w, r)
    })
}

func (cfg *apiConfig) resetMetricsHandler(w http.ResponseWriter, r *http.Request) {
    cfg.fileserverHits.Store(0)
    w.WriteHeader(http.StatusOK)
}


func readinessHandler(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/plain; charset=utf-8")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
}  

func main() {
   mux := http.NewServeMux()
   apiCfg := &apiConfig{}

   fileServer := http.FileServer(http.Dir("."))   
   mux.Handle("/app/", apiCfg.middlewareMetrics(http.StripPrefix("/app", fileServer)))
   mux.HandleFunc("GET /api/healthz", readinessHandler)
   mux.HandleFunc("GET /api/metrics", apiCfg.metricsHandler)
   mux.HandleFunc("POST /api/reset", apiCfg.resetMetricsHandler)

   server := &http.Server{
    Addr: "127.0.0.1:8080", // forces WSL2 to use IPv4
    Handler: mux,
   }

   fmt.Println("Server starting...")
   err :=  server.ListenAndServe()
   if err != nil {
        fmt.Println(err)
   }
}
