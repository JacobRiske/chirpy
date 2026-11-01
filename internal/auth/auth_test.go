package auth

import (
    "testing"
    "time"

    "github.com/google/uuid"
)

func TestMakeAndValidateJWT(t *testing.T) {
    userID := uuid.New()
    secret := "super-secret-key"

    token, err := MakeJWT(userID, secret, 5*time.Minute)
    if err != nil {
        t.Fatalf("MakeJWT returned error: %v", err)
    }

    got, err := ValidateJWT(token, secret)
    if err != nil {
        t.Fatalf("ValidateJWT returned error: %v", err)
    }

    if got != userID {
        t.Fatalf("expected user ID %s, got %s", userID, got)
    }
}

func TestValidateJWTRejectsExpiredToken(t *testing.T) {
    userID := uuid.New()
    secret := "super-secret-key"

    token, err := MakeJWT(userID, secret, -time.Minute)
    if err != nil {
        t.Fatalf("MakeJWT returned error: %v", err)
    }

    if _, err = ValidateJWT(token, secret); err == nil {
        t.Fatal("expected expired token to be rejected")
    }
}

func TestValidateJWTRejectsWrongSecret(t *testing.T) {
    userID := uuid.New()
    token, err := MakeJWT(userID, "correct-secret", 5*time.Minute)
    if err != nil {
        t.Fatalf("MakeJWT returned error: %v", err)
    }

    if _, err = ValidateJWT(token, "wrong-secret"); err == nil {
        t.Fatal("expected JWT signed with wrong secret to be rejected")
    }
}
