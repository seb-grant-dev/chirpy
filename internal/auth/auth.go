package auth

import (
	"fmt"
	"errors"
	"time"
	"strings"
	"encoding/hex"
	"net/http"
	"crypto/rand"
	"github.com/google/uuid"
	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"

)

func HashPassword(password string) (string,error) {
	hash,err := argon2id.CreateHash(password, argon2id.DefaultParams)

	if err != nil {
		return "", err
	}

	return hash, nil

}

func CheckPasswordHash(password, hash string) (bool,error) {
	match, err := argon2id.ComparePasswordAndHash(password, hash)

	if err != nil {
		return false, err
	}


	return match, nil

}

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string,error) {

	signingKey := []byte(tokenSecret)

	claims := jwt.RegisteredClaims{
		Issuer: "chirpy-access",
		IssuedAt: jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
		Subject: userID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims )
	newJwtStr, err := token.SignedString(signingKey)

	return newJwtStr, err

}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claimsType := jwt.RegisteredClaims{}
	token,err := jwt.ParseWithClaims(tokenString, &claimsType, func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})

	if err != nil {
		return uuid.Nil, err
	}
	
	userIDString, err := token.Claims.GetSubject()
	if err != nil {
		return uuid.Nil, err
	}

	issuer, err := token.Claims.GetIssuer()
	if err != nil {
		return uuid.Nil, err
	}

	if issuer != "chirpy-access" {
		return uuid.Nil, errors.New("invalid issuer")
	}


	user_id, err := uuid.Parse(userIDString)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID: %w\n", err)
	}
	return user_id, nil
}

func GetBearerToken(headers http.Header) (string,error) {
	auth := headers.Get("Authorization")
	if len(auth) == 0 {
		return "", fmt.Errorf("Not logged in")
	}

	return strings.TrimPrefix(auth,"Bearer "), nil
}

func MakeRefreshToken() string {
	key := make([]byte,32)
	_,err := rand.Read(key)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(key)
}

func GetAPIKey(headers http.Header) (string,error) {
	auth := headers.Get("Authorization")
	if len(auth) == 0 {
		return "", fmt.Errorf("API Key not provided")
	}
	return strings.TrimPrefix(auth,"ApiKey "),nil
}
