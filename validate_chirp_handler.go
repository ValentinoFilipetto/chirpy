package main

import (
  "net/http"
  "encoding/json"
  "log"
)


func (cfg *apiConfig) ValidateChirpHandler(w http.ResponseWriter, r *http.Request) {
   type parameters struct {
       Body string `json:"body"`
   }

   type returnVals struct {
       Valid bool `json:"valid"`
   }

   type errorVals struct {
       Error string `json:"error"`
   }

   respBody := returnVals{}
   errorBody := errorVals{}

   decoder := json.NewDecoder(r.Body)
   params := parameters{}
   err := decoder.Decode(&params)
   if err != nil {
     log.Printf("Error decoding parameters: %s", err)
     w.WriteHeader(500)
     return
   }

   w.Header().Set("Content-Type", "application/json")

   if len(params.Body) <= 140 {
        respBody.Valid = true
        dat, err := json.Marshal(respBody)

        if err != nil {
             log.Printf("Error marshalling JSON")
             w.WriteHeader(500)
             return
        }
        w.WriteHeader(200)
        w.Write(dat)
   } else {
        errorBody.Error = "Chirp is too long"
        dat, err := json.Marshal(errorBody)

        if err != nil {
                log.Printf("Error marshalling JSON")
                w.WriteHeader(500)
                return
        }
        w.WriteHeader(400)
        w.Write(dat)
   }
}
