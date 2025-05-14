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

	// Determine JWT expiration time
	const jwtExpirationTime = 3600

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

	jwt, err := auth.MakeJWT(user.ID, cfg.JWT_SECRET, time.Duration(jwtExpirationTime)*time.Second)

	if err != nil {
		log.Printf("Error creating JWT: %s", err)
		w.WriteHeader(500)
		return
	}

	refreshToken, refreshTokenErr := auth.MakeRefreshToken()

	if refreshTokenErr != nil {
		log.Printf("Error creating refresh token: %s", refreshTokenErr)
		w.WriteHeader(500)
		return
	}

	refreshTokenParams := database.CreateRefreshTokenParams{
		Token:  refreshToken,
		UserID: user.ID,
	}

	_, err = cfg.DB.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams(refreshTokenParams))

	if err != nil {
		log.Printf("Error storing refresh token: %s", err)
		w.WriteHeader(500)
		return
	}

	respBody := struct {
		ID           uuid.UUID `json:"id"`
		CreatedAt    time.Time `json:"created_at"`
		UpdatedAt    time.Time `json:"updated_at"`
		Email        string    `json:"email"`
		Token        string    `json:"token"`
		RefreshToken string    `json:"refresh_token"`
	}{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        jwt,
		RefreshToken: refreshToken,
	}

	respondWithJSON(w, 200, respBody)

}

func (cfg *apiConfig) RefreshTokenHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		log.Printf("Error getting token from the header: %s", err)
		w.WriteHeader(500)
		return
	}

	refreshToken, err := cfg.DB.GetRefreshToken(r.Context(), token)

	if err != nil {
		respondWithError(w, 401, "Refresh token cannot be found in the database")
		return
	}

	if time.Now().After(refreshToken.ExpiresAt) {
		respondWithError(w, 401, "Refresh token expired")
		return
	}

	if refreshToken.RevokedAt.Valid {
		// If RevokedAt has a valid (non-NULL) timestamp, the token is revoked
		respondWithError(w, 401, "Refresh token has been revoked")
		return
	}

	user, err := cfg.DB.GetUserFromRefreshToken(r.Context(), refreshToken.UserID)

	if err != nil {
		respondWithError(w, 404, "Cannot find user based on the refresh token")
		return
	}

	jwt, err := auth.MakeJWT(user.ID, cfg.JWT_SECRET, time.Duration(3600)*time.Second)

	if err != nil {
		respondWithError(w, 500, "Error creating new JWT")
		return
	}

	respBody := struct {
		Token string `json:"token"`
	}{
		Token: jwt,
	}

	respondWithJSON(w, 200, respBody)

}

func (cfg *apiConfig) RevokeTokenHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)

	if err != nil {
		log.Printf("Error getting token from the header: %s", err)
		w.WriteHeader(500)
		return
	}

	err = cfg.DB.RevokeRefreshToken(r.Context(), token)

	if err != nil {
		respondWithError(w, 500, "Cannot update refresh token in database")
		return
	}

	w.WriteHeader(204)
}
