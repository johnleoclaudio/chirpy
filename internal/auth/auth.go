package auth

import (
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

func MakeJWT(user uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject:   user.String(),
	})

	jwtToken, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return jwtToken, err
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	jwtToken, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	userID := jwtToken.Claims.(*jwt.RegisteredClaims).Subject
	userUUID, err := uuid.FromBytes([]byte(userID))
	if err != nil {
		return userUUID, err
	}

	return userUUID, nil
}
