package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPassword(password string) (string, error) {
	return argon2id.CreateHash(password, argon2id.DefaultParams)
}

func CheckPassword(password, hashedPassword string) (bool, error) {
	return argon2id.ComparePasswordAndHash(password, hashedPassword)
}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   userID.String(),
	})

	jwtToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return jwtToken, nil
}

func ValidateJWT(tokenString, tokenSecret string) (string, error) {
	jwtToken, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		log.Printf("failed to validate JWT: %v, %s", err, tokenString)
		return "", err
	}

	userID := jwtToken.Claims.(*jwt.RegisteredClaims).Subject
	return userID, nil
}

func GetBearerToken(headers http.Header) (string, error) {
	authHeader := headers.Get("Authorization")
	if authHeader == "" {
		return "", fmt.Errorf("auth header missing")
	}

	token := strings.Split(authHeader, " ")
	if len(token) != 2 || strings.ToLower(token[0]) != "bearer" {
		return "", fmt.Errorf("malformed auth header")
	}

	return token[1], nil
}

func GetRefreshTokenHeader(headers http.Header) (string, error) {
	refresh := headers.Get("Authorization")
	if refresh == "" {
		return "", fmt.Errorf("refresh header missing")
	}

	token := strings.Split(refresh, " ")
	if len(token) != 2 || strings.ToLower(token[0]) != "bearer" {
		return "", fmt.Errorf("malformed auth header")
	}

	return token[1], nil
}

func MakeRefreshToken() (string, error) {
	randomBytes := make([]byte, 32)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	refreshToken := hex.EncodeToString(randomBytes)
	return refreshToken, nil
}

func GetAPIKey(headers http.Header) (string, error) {
	apiKey := headers.Get("Authorization")
	if apiKey == "" {
		return "", fmt.Errorf("auth header missing")
	}

	key := strings.Split(apiKey, " ")
	if len(key) != 2 || strings.ToLower(key[0]) != "apikey" {
		return "", fmt.Errorf("malformed auth header")
	}

	return key[1], nil
}
