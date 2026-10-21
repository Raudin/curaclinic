// Package token issues and verifies JWT access tokens and opaque refresh
// tokens.
package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims are the identity-only claims stored in access tokens. Permissions are
// always loaded from the database/cache so revocations take effect immediately.
type Claims struct {
	SessionID uuid.UUID `json:"sid"`
	jwt.RegisteredClaims
}

// Service signs and verifies access tokens.
type Service struct {
	secret    []byte
	issuer    string
	audience  string
	accessTTL time.Duration
}

// New creates a token service.
func New(secret, issuer, audience string, accessTTL time.Duration) *Service {
	return &Service{
		secret:    []byte(secret),
		issuer:    issuer,
		audience:  audience,
		accessTTL: accessTTL,
	}
}

// AccessTTL returns the configured access-token lifetime.
func (s *Service) AccessTTL() time.Duration { return s.accessTTL }

// IssueAccess creates a signed access token for a user and session.
func (s *Service) IssueAccess(userID, sessionID uuid.UUID, now time.Time) (string, time.Time, error) {
	expiresAt := now.Add(s.accessTTL).UTC()
	claims := Claims{
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			Issuer:    s.issuer,
			Audience:  jwt.ClaimStrings{s.audience},
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now.UTC()),
			NotBefore: jwt.NewNumericDate(now.Add(-5 * time.Second).UTC()),
			ID:        uuid.NewString(),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.secret)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}
	return signed, expiresAt, nil
}

// Verify parses and validates an access token.
func (s *Service) Verify(raw string) (*Claims, error) {
	claims := &Claims{}
	parsed, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, fmt.Errorf("unexpected signing method %q", t.Method.Alg())
		}
		return s.secret, nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(s.issuer),
		jwt.WithAudience(s.audience),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(30*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("verify access token: %w", err)
	}
	if !parsed.Valid {
		return nil, fmt.Errorf("access token is invalid")
	}
	if _, err := uuid.Parse(claims.Subject); err != nil {
		return nil, fmt.Errorf("access token subject is invalid")
	}
	if claims.SessionID == uuid.Nil {
		return nil, fmt.Errorf("access token session is missing")
	}
	return claims, nil
}

// GenerateRefreshToken returns a 256-bit opaque token.
func GenerateRefreshToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// HashRefreshToken returns a stable SHA-256 hex digest for storage.
func HashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
