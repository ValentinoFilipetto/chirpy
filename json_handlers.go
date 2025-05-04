package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type errorVals struct {
	Error string `json:"error"`
}

type returnVals struct {
	Cleaned_body string `json:"cleaned_body"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
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
	errorBody.Error = msg
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

func (cfg *apiConfig) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email string `json:"email"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	user, err := cfg.DB.CreateUser(r.Context(), params.Email)

	if err != nil {
		log.Printf("Error creating user: %s", err)
		w.WriteHeader(500)
		return
	}

	respBody := User{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
	}

	respondWithJSON(w, 201, respBody)
}
