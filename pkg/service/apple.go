package service

import (
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Sovpalo/sovpalo-backend/pkg/model"
	"github.com/jackc/pgx/v5"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

const (
	appleJWKSURL = "https://appleid.apple.com/auth/keys"
	appleIssuer  = "https://appleid.apple.com"
)

type validatedAppleIdentity struct {
	AppleUserID string
	Email       *string
}

type appleIdentityValidator interface {
	Validate(identityToken string, nonce *string) (validatedAppleIdentity, error)
}

type appleJWKSValidator struct {
	client    *http.Client
	clientID  string
	mu        sync.Mutex
	keys      map[string]*rsa.PublicKey
	fetchedAt time.Time
	ttl       time.Duration
}

func newAppleJWKSValidator(clientID string, client *http.Client) *appleJWKSValidator {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &appleJWKSValidator{
		client:   client,
		clientID: clientID,
		keys:     make(map[string]*rsa.PublicKey),
		ttl:      time.Hour,
	}
}

func (s *AuthService) SignInApple(input model.AppleSignInInput) (string, error) {
	identity, err := s.validateAppleIdentity(input.IdentityToken, input.Nonce)
	if err != nil {
		return "", err
	}

	user, err := s.repo.GetUserByAppleID(identity.AppleUserID)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return "", err
		}

		username, usernameErr := s.generateAppleUsername(input, identity.AppleUserID, identity.Email)
		if usernameErr != nil {
			return "", usernameErr
		}

		appleUserID := identity.AppleUserID
		userID, createErr := s.repo.CreateUser(model.User{
			AppleUserID: &appleUserID,
			Username:    username,
		})
		if createErr != nil {
			return "", createErr
		}

		return s.generateTokenByUserID(int64(userID))
	}

	return s.generateTokenByUserID(user.ID)
}

func (s *AuthService) validateAppleIdentity(identityToken string, nonce *string) (validatedAppleIdentity, error) {
	clientID := strings.TrimSpace(os.Getenv("APPLE_CLIENT_ID"))
	if clientID == "" {
		clientID = strings.TrimSpace(os.Getenv("APNS_BUNDLE_ID"))
	}
	if clientID == "" {
		return validatedAppleIdentity{}, errors.New("APPLE_CLIENT_ID not set")
	}

	validator := s.appleValidator
	if validator == nil {
		validator = newAppleJWKSValidator(clientID, nil)
	}

	return validator.Validate(identityToken, nonce)
}

func (v *appleJWKSValidator) Validate(identityToken string, nonce *string) (validatedAppleIdentity, error) {
	identityToken = strings.TrimSpace(identityToken)
	if identityToken == "" {
		return validatedAppleIdentity{}, ErrInvalidAppleAuth
	}

	parser := jwtv5.NewParser(jwtv5.WithValidMethods([]string{jwtv5.SigningMethodRS256.Alg()}))
	claims := &appleIdentityClaims{}
	token, err := parser.ParseWithClaims(identityToken, claims, func(token *jwtv5.Token) (interface{}, error) {
		kid, ok := token.Header["kid"].(string)
		if !ok || kid == "" {
			return nil, ErrInvalidAppleAuth
		}
		return v.publicKey(kid)
	})
	if err != nil {
		return validatedAppleIdentity{}, ErrInvalidAppleAuth
	}
	if !token.Valid {
		return validatedAppleIdentity{}, ErrInvalidAppleAuth
	}

	if claims.Issuer != appleIssuer {
		return validatedAppleIdentity{}, ErrInvalidAppleAuth
	}
	if !appleAudienceContains(claims.Audience, v.clientID) {
		return validatedAppleIdentity{}, ErrInvalidAppleAuth
	}
	if claims.ExpiresAt == nil || time.Now().After(claims.ExpiresAt.Time) {
		return validatedAppleIdentity{}, ErrInvalidAppleAuth
	}
	if strings.TrimSpace(claims.Subject) == "" {
		return validatedAppleIdentity{}, ErrInvalidAppleAuth
	}
	if err := verifyAppleNonce(claims.Nonce, nonce); err != nil {
		return validatedAppleIdentity{}, err
	}

	var email *string
	if claims.Email != "" {
		value := strings.TrimSpace(claims.Email)
		if value != "" {
			email = &value
		}
	}

	return validatedAppleIdentity{
		AppleUserID: claims.Subject,
		Email:       email,
	}, nil
}

func (v *appleJWKSValidator) publicKey(kid string) (*rsa.PublicKey, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	if len(v.keys) > 0 && time.Since(v.fetchedAt) < v.ttl {
		if key, ok := v.keys[kid]; ok {
			return key, nil
		}
	}

	if err := v.refreshKeys(); err != nil {
		return nil, err
	}

	key, ok := v.keys[kid]
	if !ok {
		return nil, ErrInvalidAppleAuth
	}
	return key, nil
}

func (v *appleJWKSValidator) refreshKeys() error {
	req, err := http.NewRequest(http.MethodGet, appleJWKSURL, nil)
	if err != nil {
		return err
	}

	resp, err := v.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("apple jwks request failed with status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}

	var payload struct {
		Keys []appleJWK `json:"keys"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return err
	}

	keys := make(map[string]*rsa.PublicKey, len(payload.Keys))
	for _, key := range payload.Keys {
		if key.Kty != "RSA" || key.Kid == "" || key.N == "" || key.E == "" {
			continue
		}
		publicKey, err := parseAppleRSAPublicKey(key.N, key.E)
		if err != nil {
			continue
		}
		keys[key.Kid] = publicKey
	}
	if len(keys) == 0 {
		return ErrInvalidAppleAuth
	}

	v.keys = keys
	v.fetchedAt = time.Now()
	return nil
}

type appleJWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type appleIdentityClaims struct {
	jwtv5.RegisteredClaims
	Email string `json:"email"`
	Nonce string `json:"nonce"`
}

func parseAppleRSAPublicKey(nB64, eB64 string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nB64)
	if err != nil {
		return nil, err
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(eB64)
	if err != nil {
		return nil, err
	}

	n := new(big.Int).SetBytes(nBytes)
	e := new(big.Int).SetBytes(eBytes)
	if !e.IsInt64() {
		return nil, errors.New("invalid rsa exponent")
	}

	return &rsa.PublicKey{
		N: n,
		E: int(e.Int64()),
	}, nil
}

func appleAudienceContains(audience jwtv5.ClaimStrings, clientID string) bool {
	for _, item := range audience {
		if item == clientID {
			return true
		}
	}
	return false
}

func verifyAppleNonce(tokenNonce string, nonce *string) error {
	if nonce == nil || strings.TrimSpace(*nonce) == "" {
		return nil
	}
	if strings.TrimSpace(tokenNonce) == "" {
		return ErrInvalidAppleAuth
	}

	hash := sha256.Sum256([]byte(strings.TrimSpace(*nonce)))
	expected := hex.EncodeToString(hash[:])
	if !strings.EqualFold(tokenNonce, expected) {
		return ErrInvalidAppleAuth
	}
	return nil
}

func (s *AuthService) generateAppleUsername(input model.AppleSignInInput, appleUserID string, tokenEmail *string) (string, error) {
	candidates := make([]string, 0, 4)

	fullName := ""
	if input.GivenName != nil {
		fullName = strings.TrimSpace(*input.GivenName)
	}
	if input.FamilyName != nil && strings.TrimSpace(*input.FamilyName) != "" {
		if fullName != "" {
			fullName += "_"
		}
		fullName += strings.TrimSpace(*input.FamilyName)
	}
	if username := sanitizeTelegramUsername(fullName); username != "" {
		candidates = append(candidates, username)
	}

	for _, emailPtr := range []*string{tokenEmail, input.Email} {
		if emailPtr == nil {
			continue
		}
		local := emailLocalPartForUsername(*emailPtr)
		if local == "" {
			continue
		}
		if username := sanitizeTelegramUsername(local); username != "" {
			candidates = append(candidates, username)
		}
	}

	safeID := strings.ToLower(strings.ReplaceAll(appleUserID, ".", "_"))
	if len(safeID) > 12 {
		safeID = safeID[:12]
	}
	candidates = append(candidates, "apple_"+safeID)

	for _, candidate := range candidates {
		username, err := s.ensureUniqueUsername(candidate, 0)
		if err != nil {
			return "", err
		}
		if username != "" {
			return username, nil
		}
	}

	return "", ErrUsernameAlreadyExists
}

func emailLocalPartForUsername(email string) string {
	email = strings.TrimSpace(strings.ToLower(email))
	at := strings.LastIndex(email, "@")
	if at <= 0 || at == len(email)-1 {
		return ""
	}
	local := email[:at]
	if plus := strings.Index(local, "+"); plus >= 0 {
		local = local[:plus]
	}
	return strings.TrimSpace(local)
}
