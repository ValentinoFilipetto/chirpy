package auth

import (
	"net/http"
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "password123"
	hashedPassword, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	if hashedPassword == password {
		t.Fatal("Hashed password should not be the same as the original password")
	}
}

func TestGetBearerToken(t *testing.T) {
	expectedToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"
	headers := http.Header{
		"Authorization": []string{"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"},
	}

	token, err := GetBearerToken(headers)
	if err != nil {
		t.Fatalf("authorization is not present in header")
	}

	if token != expectedToken {
		t.Fatalf("Expected token %s, got %s", expectedToken, token)
	}
}
