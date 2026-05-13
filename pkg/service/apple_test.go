package service

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

type appleIdentityValidatorStub struct {
	identity validatedAppleIdentity
	err      error
}

func (s appleIdentityValidatorStub) Validate(identityToken string, nonce *string) (validatedAppleIdentity, error) {
	return s.identity, s.err
}

func TestAuthServiceSignInAppleCreatesUser(t *testing.T) {
	t.Setenv("JWT_SECRET", "jwt-secret")
	t.Setenv("APPLE_CLIENT_ID", "com.sovpalo.test")

	repo := &authRepoStub{
		userByAppleID:  map[string]model.User{},
		takenUsernames: map[string]bool{"ivan": true},
	}
	svc := NewAuthService(repo)
	svc.appleValidator = appleIdentityValidatorStub{
		identity: validatedAppleIdentity{AppleUserID: "001234.abcdef"},
	}

	givenName := "Ivan"
	familyName := "Petrov"
	token, err := svc.SignInApple(model.AppleSignInInput{
		IdentityToken: "token",
		GivenName:     &givenName,
		FamilyName:    &familyName,
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Fatal("expected token")
	}
	if len(repo.createdUsers) != 1 {
		t.Fatalf("expected 1 created user, got %d", len(repo.createdUsers))
	}
	if repo.createdUsers[0].Username != "ivan_petrov" {
		t.Fatalf("expected username ivan_petrov, got %s", repo.createdUsers[0].Username)
	}
	if repo.createdUsers[0].AppleUserID == nil || *repo.createdUsers[0].AppleUserID != "001234.abcdef" {
		t.Fatalf("unexpected apple user id %#v", repo.createdUsers[0].AppleUserID)
	}
}

func TestAuthServiceSignInAppleReturnsExistingUserToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "jwt-secret")
	t.Setenv("APPLE_CLIENT_ID", "com.sovpalo.test")

	repo := &authRepoStub{
		userByAppleID: map[string]model.User{
			"001234.abcdef": {ID: 42, Username: "apple_user"},
		},
	}
	svc := NewAuthService(repo)
	svc.appleValidator = appleIdentityValidatorStub{
		identity: validatedAppleIdentity{AppleUserID: "001234.abcdef"},
	}

	token, err := svc.SignInApple(model.AppleSignInInput{IdentityToken: "token"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Fatal("expected token")
	}
	if len(repo.createdUsers) != 0 {
		t.Fatalf("expected no created users, got %d", len(repo.createdUsers))
	}
}

func TestAuthServiceSignInAppleRejectsInvalidToken(t *testing.T) {
	t.Setenv("JWT_SECRET", "jwt-secret")
	t.Setenv("APPLE_CLIENT_ID", "com.sovpalo.test")

	svc := NewAuthService(&authRepoStub{})
	svc.appleValidator = appleIdentityValidatorStub{err: ErrInvalidAppleAuth}

	_, err := svc.SignInApple(model.AppleSignInInput{IdentityToken: "bad-token"})
	if !errors.Is(err, ErrInvalidAppleAuth) {
		t.Fatalf("expected %v, got %v", ErrInvalidAppleAuth, err)
	}
}

func TestVerifyAppleNonce(t *testing.T) {
	raw := "session-nonce"
	hash := sha256Hex([]byte(raw))

	if err := verifyAppleNonce(hash, &raw); err != nil {
		t.Fatalf("expected valid nonce, got %v", err)
	}
	if err := verifyAppleNonce("bad", &raw); !errors.Is(err, ErrInvalidAppleAuth) {
		t.Fatalf("expected %v, got %v", ErrInvalidAppleAuth, err)
	}
}

func TestAppleJWKSValidatorValidate(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}

	kid := "test-key"
	token := jwtv5.NewWithClaims(jwtv5.SigningMethodRS256, &appleIdentityClaims{
		RegisteredClaims: jwtv5.RegisteredClaims{
			Issuer:    appleIssuer,
			Audience:  jwtv5.ClaimStrings{"com.sovpalo.test"},
			Subject:   "001234.abcdef",
			ExpiresAt: jwtv5.NewNumericDate(time.Now().Add(time.Hour)),
		},
	})
	token.Header["kid"] = kid
	signed, err := token.SignedString(privateKey)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}

	validator := newAppleJWKSValidator("com.sovpalo.test", nil)
	validator.keys = map[string]*rsa.PublicKey{kid: &privateKey.PublicKey}
	validator.fetchedAt = time.Now()

	identity, err := validator.Validate(signed, nil)
	if err != nil {
		t.Fatalf("expected valid token, got %v", err)
	}
	if identity.AppleUserID != "001234.abcdef" {
		t.Fatalf("unexpected apple user id %q", identity.AppleUserID)
	}
}

func sha256Hex(value []byte) string {
	sum := sha256.Sum256(value)
	return hex.EncodeToString(sum[:])
}
