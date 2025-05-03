package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type errorVals struct {
	Error string `json:"error"`
}

type returnVals struct {
	Cleaned_body string `json:"cleaned_body"`
}

// Handler for JSON responses
func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	dat, err := json.Marshal(payload)

	if err != nil {
		log.Printf("Error marshalling JSON")
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

// Handler for error responses
func respondWithError(w http.ResponseWriter, code int, msg string) {
	errorBody := errorVals{}
	errorBody.Error = "Chirp is too long"
	dat, err := json.Marshal(errorBody)

	if err != nil {
		log.Printf("Error marshalling JSON")
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(code)
	w.Write(dat)
}

func badWorldReplacement(str string) string {
	profaneWords := []string{"kerfuffle", "sharbert", "fornax"}
	words := strings.Split(str, " ")

	for i, word := range words {
		for _, profaneWord := range profaneWords {
			lowercaseWord := strings.ToLower(word)
			if lowercaseWord == profaneWord {
				words[i] = "****"
			}
		}
	}

	return strings.Join(words, " ")
}

func (cfg *apiConfig) ValidateChirpHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Body string `json:"body"`
	}

	respBody := returnVals{}

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
		params.Body = badWorldReplacement(params.Body)
		respBody.Cleaned_body = params.Body
		respondWithJSON(w, 200, respBody)
	} else {
		respondWithError(w, 400, "Chirp is too long")
	}
}
