package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	res, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return "", err
	}

	return string(res), nil
}

func CheckPasswordHash(hash, password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return err
	}

	return nil
}

// MakeJWT creates a new JWT token for the given user ID and secret.
// In particular, in the token we put a user ID and an expiration date.
// We further make sure the token is signed with the secret
func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	secretKey := []byte(tokenSecret)

	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
		ExpiresAt: jwt.NewNumericDate(time.Now().UTC().Add(expiresIn)),
		Subject:   userID.String(),
	}
	res := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwt, err := res.SignedString(secretKey)

	if err != nil {
		return "", err
	}

	return jwt, nil
}

// ValidateJWT checks if the token is valid and not expired.
// In particular, we check if the token was signed with HS256 signing method.
// If everything is ok, we return the user ID from the token.
func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}

	keyFunc := func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(tokenSecret), nil
	}

	token, err := jwt.ParseWithClaims(tokenString, claims, keyFunc)
	if err != nil {
		fmt.Printf("Token validation error: %v\n", err)
		return uuid.Nil, err
	}

	userIDStr := token.Claims.(*jwt.RegisteredClaims).Subject
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authorizationString := headers.Get("Authorization")

	if authorizationString == "" {
		return "", fmt.Errorf("authorization is not present in header")
	}

	return strings.TrimPrefix(authorizationString, "Bearer "), nil

}

func GetApiKey(headers http.Header) (string, error) {
	authorizationString := headers.Get("Authorization")

	if authorizationString == "" {
		return "", fmt.Errorf("authorization is not present in header")
	}

	return strings.TrimPrefix(authorizationString, "ApiKey "), nil

}

func MakeRefreshToken() (string, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)

	if err != nil {
		return "", err
	}

	token := hex.EncodeToString(key)
	return token, nil
}
