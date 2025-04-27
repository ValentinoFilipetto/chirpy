package main

import (
   "net/http"
   "fmt"
  )

func main() {
   mux := http.NewServeMux()

   mux.Handle("/", http.FileServer(http.Dir(".")))

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
