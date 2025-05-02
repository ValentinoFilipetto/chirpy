package main

import (
   "net/http"
   "fmt"
  )

func readinessHandler(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/plain; charset=utf-8")
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("OK"))
}  

func main() {
   mux := http.NewServeMux()

   fileServer := http.FileServer(http.Dir("."))   
   mux.Handle("/app/", http.StripPrefix("/app", fileServer))
   mux.HandleFunc("/healthz", readinessHandler)

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
