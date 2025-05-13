package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/ValentinoFilipetto/chirpy/internal/auth"
	"github.com/ValentinoFilipetto/chirpy/internal/database"
	"github.com/google/uuid"
)

type errorVals struct {
	Error string `json:"error"`
}

type User struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

type Chirp struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	UserID    uuid.UUID `json:"user_id"`
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

func (cfg *apiConfig) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	hashedPassword, err := auth.HashPassword(params.Password)

	if err != nil {
		log.Printf("Error hashing password: %s", err)
		w.WriteHeader(500)
		return
	}

	createUserParams := database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPassword,
	}

	w.Header().Set("Content-Type", "application/json")

	user, err := cfg.DB.CreateUser(r.Context(), createUserParams)

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

func (cfg *apiConfig) AddChirpHandler(w http.ResponseWriter, r *http.Request) {

	type parameters struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}

	token, tokenError := auth.GetBearerToken(r.Header)

	if tokenError != nil {
		log.Printf("Error getting bearer token: %s", tokenError)
		w.WriteHeader(401)
		return
	}

	userIDFromJWT, err := auth.ValidateJWT(token, cfg.JWT_SECRET)

	if err != nil {
		log.Printf("Error validating JWT: %s", err)
		w.WriteHeader(401)
		return
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err = decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if len(params.Body) <= 140 {
		params.Body = badWorldReplacement(params.Body)
	} else {
		respondWithError(w, 400, "Chirp is too long")
		return
	}

	chirpParams := database.CreateChirpParams{
		Body:   params.Body,
		UserID: userIDFromJWT,
	}

	chirp, err := cfg.DB.CreateChirp(r.Context(), chirpParams)

	if err != nil {
		log.Printf("Error creating chirp: %s", err)
		w.WriteHeader(500)
		return
	}

	respBody := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    userIDFromJWT,
	}

	respondWithJSON(w, 201, respBody)
}

func (cfg *apiConfig) GetChirpsHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	chirps, err := cfg.DB.GetAllChirps(r.Context())

	if err != nil {
		log.Printf("Error retrieving chirps from database: %s", err)
		w.WriteHeader(500)
		return
	}

	respBody := make([]Chirp, len(chirps))
	for i, chirp := range chirps {
		respBody[i] = Chirp{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		}
	}

	dat, err := json.Marshal(respBody)

	if err != nil {
		log.Printf("Error marshalling JSON")
		w.WriteHeader(500)
		return
	}
	w.WriteHeader(200)
	w.Write(dat)
}

func (cfg *apiConfig) GetChirpHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	chirpID, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		log.Printf("Invalid chirpID: %s", err)
		respondWithError(w, 400, "Invalid chirpID")
		return
	}

	chirp, err := cfg.DB.GetChirp(r.Context(), chirpID)

	if err != nil {
		log.Printf("Error retrieving chirp from database: %s", err)
		w.WriteHeader(404)
		return
	}

	respBody := Chirp{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	respondWithJSON(w, 200, respBody)
}

func (cfg *apiConfig) LoginUserHandler(w http.ResponseWriter, r *http.Request) {
	type parameters struct {
		Email            string `json:"email"`
		Password         string `json:"password"`
		ExpiresInSeconds int    `json:"expires_in_seconds"`
	}

	decoder := json.NewDecoder(r.Body)
	params := parameters{}
	err := decoder.Decode(&params)
	if err != nil {
		log.Printf("Error decoding parameters: %s", err)
		w.WriteHeader(500)
		return
	}

	// Determine JWT expiration time
	const defaultExpiration = 3600
	jwtExpirationTime := time.Duration(defaultExpiration) * time.Second
	if params.ExpiresInSeconds > 0 && params.ExpiresInSeconds <= defaultExpiration {
		jwtExpirationTime = time.Duration(params.ExpiresInSeconds) * time.Second
	}

	w.Header().Set("Content-Type", "application/json")

	user, err := cfg.DB.GetUserByEmail(r.Context(), params.Email)

	if err != nil {
		log.Printf("Error retrieving user from database: %s", err)
		w.WriteHeader(404)
		return
	}

	err = auth.CheckPasswordHash(user.HashedPassword, params.Password)

	if err != nil {
		log.Printf("Incorrect email or password")
		w.WriteHeader(401)
		return
	}

	jwt, err := auth.MakeJWT(user.ID, cfg.JWT_SECRET, jwtExpirationTime)

	if err != nil {
		log.Printf("Error creating JWT: %s", err)
		w.WriteHeader(500)
		return
	}

	respBody := struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
		Token     string    `json:"token"`
	}{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     jwt,
	}

	respondWithJSON(w, 200, respBody)

}
