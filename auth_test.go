package main

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestValidateToken(t *testing.T) {
	jwtSecret = []byte("testsecretjwt12345678901234567890")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Minute).Unix(),
	})
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		t.Fatalf("Failed to sign token: %v", err)
	}
	if !validateToken(tokenString) {
		t.Error("validateToken should return true for valid token")
	}

	// Expired token
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(-time.Minute).Unix(),
	})
	expiredString, _ := expiredToken.SignedString(jwtSecret)
	if validateToken(expiredString) {
		t.Error("validateToken should return false for expired token")
	}

	// Invalid token
	if validateToken("invalid.token.string") {
		t.Error("validateToken should return false for invalid token")
	}
}

func TestValidateToken_Malformed(t *testing.T) {
	jwtSecret = []byte("testsecretjwt12345678901234567890")
	// Token with wrong signature
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Minute).Unix(),
	})
	tokenString, _ := token.SignedString([]byte("wrongsecret"))
	if validateToken(tokenString) {
		t.Error("validateToken should return false for token with wrong signature")
	}
	// Token with wrong algorithm
	token = jwt.NewWithClaims(jwt.SigningMethodNone, jwt.MapClaims{
		"exp": time.Now().Add(time.Minute).Unix(),
	})
	tokenString, _ = token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if validateToken(tokenString) {
		t.Error("validateToken should return false for token with wrong algorithm")
	}
}

func TestValidateToken_Expired(t *testing.T) {
	jwtSecret = []byte("testsecretjwt12345678901234567890")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(-time.Hour).Unix(), // expired 1 hour ago
	})
	tokenString, _ := token.SignedString(jwtSecret)
	if validateToken(tokenString) {
		t.Error("validateToken should return false for expired token")
	}
}

func TestCheckAuth(t *testing.T) {
	jwtSecret = []byte("testsecretjwt12345678901234567890")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"exp": time.Now().Add(time.Minute).Unix(),
	})
	tokenString, _ := token.SignedString(jwtSecret)

	r := httptest.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Bearer "+tokenString)
	if !checkAuth(r) {
		t.Error("checkAuth should return true for valid token")
	}

	r.Header.Set("Authorization", "Bearer invalidtoken")
	if checkAuth(r) {
		t.Error("checkAuth should return false for invalid token")
	}
}
