package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func HashPass(s string) (string, error) {
	return argon2id.CreateHash(s, argon2id.DefaultParams)
}

func CheckPass(s, hashed string) (bool, error) {
	return argon2id.ComparePasswordAndHash(s, hashed)
}

func MakeJWT(userID uuid.UUID, tokenS string, expire time.Duration) (string, error) {
	if tokenS == "" {
		return "", errors.New("token secret is empty")
	}

	claims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expire)),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(tokenS))
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	if tokenString == "" {
		return uuid.Nil, errors.New("token string is empty")
	}
	if tokenSecret == "" {
		return uuid.Nil, errors.New("token secret is empty")
	}

	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Method.Alg())
		}
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok || !token.Valid {
		return uuid.Nil, errors.New("invalid token claims")
	}
	if claims.Subject == "" {
		return uuid.Nil, errors.New("missing subject in token")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, err
	}

	return userID, nil
}

func GetBearToken(headers http.Header) (string, error) {
	var (
		t   string
		err error
	)

	t = headers.Get("Authorization")

	if !strings.Contains(t, "Bearer") {
		return "", errors.New("un-authorized user")
	}

	t = strings.Split(t, " ")[1]

	return t, err
}

func MakeRefreshToken() string {
	key := make([]byte, 32)
	rand.Read(key)
	return hex.EncodeToString(key)
}

func GetApiToken(headers http.Header) (string, error) {

	var (
		t   string
		err error
	)

	hVal := headers.Values("Authorization")

	if len(hVal) == 0 {
		return "", errors.New("no header data")
	}

	for _, val := range hVal {
		if strings.Contains(val, "ApiKey") {
			t = val
		}
	}

	if t == "" {
		return "", errors.New("no api key")
	}

	t = strings.Split(t, " ")[1]

	return t, err
}
