// Package auth verifies bearer tokens and carries the caller's email in ctx.
package auth

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/kritpi/scaffold-golang/internal/core/domain"
)

// Config locates the token signer's public key and names the expected issuer and audience.
type Config struct {
	PublicKeyFile string
	Issuer        string
	Audience      string
}

// Verifier accepts ES256 tokens from Config.Issuer for Config.Audience.
type Verifier struct {
	key    *ecdsa.PublicKey
	parser *jwt.Parser
}

// New reads and parses the public key.
func New(cfg Config) (*Verifier, error) {
	pem, err := os.ReadFile(cfg.PublicKeyFile)
	if err != nil {
		return nil, fmt.Errorf("auth: read public key: %w", err)
	}
	key, err := jwt.ParseECPublicKeyFromPEM(pem)
	if err != nil {
		return nil, fmt.Errorf("auth: parse public key %s: %w", cfg.PublicKeyFile, err)
	}
	return &Verifier{key: key, parser: jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodES256.Alg()}),
		jwt.WithIssuer(cfg.Issuer),
		jwt.WithAudience(cfg.Audience),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(30*time.Second),
	)}, nil
}

type claims struct {
	jwt.RegisteredClaims
	Email string `json:"email"`
}

// Actor verifies token and returns its email claim. ErrUnauthenticated on any failure.
func (v *Verifier) Actor(token string) (string, error) {
	var c claims
	if _, err := v.parser.ParseWithClaims(token, &c, func(*jwt.Token) (any, error) { return v.key, nil }); err != nil {
		return "", fmt.Errorf("%w: %v", domain.ErrUnauthenticated, err)
	}
	if c.Email == "" {
		return "", fmt.Errorf("%w: token has no email claim", domain.ErrUnauthenticated)
	}
	return c.Email, nil
}

type actorKey struct{}

// WithActor returns ctx carrying the verified email.
func WithActor(ctx context.Context, email string) context.Context {
	return context.WithValue(ctx, actorKey{}, email)
}

// ActorFrom returns the verified email, or "" when ctx carries none.
func ActorFrom(ctx context.Context) string {
	email, _ := ctx.Value(actorKey{}).(string)
	return email
}
