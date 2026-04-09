package middleware

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
)

type CloudflareAccessVerifier struct {
	jwks      *keyfunc.JWKS
	issuer    string
	audiences []string
}

func NewCloudflareAccessVerifier(teamDomain string, audiences []string) (*CloudflareAccessVerifier, error) {
	if teamDomain == "" {
		return nil, errors.New("cf access: team domain is required")
	}
	if len(audiences) == 0 {
		return nil, errors.New("cf access: at least one audience is required")
	}

	issuer := strings.TrimRight(teamDomain, "/")
	certsURL := issuer + "/cdn-cgi/access/certs"

	jwks, err := keyfunc.Get(certsURL, keyfunc.Options{
		RefreshInterval:   time.Hour,
		RefreshRateLimit:  5 * time.Minute,
		RefreshTimeout:    30 * time.Second,
		RefreshUnknownKID: true,
		RefreshErrorHandler: func(err error) {
			log.Printf("cf access: jwks refresh error: %v", err)
		},
	})
	if err != nil {
		return nil, fmt.Errorf("cf access: failed to fetch JWKS from %s: %w", certsURL, err)
	}

	return &CloudflareAccessVerifier{
		jwks:      jwks,
		issuer:    issuer,
		audiences: audiences,
	}, nil
}

func (v *CloudflareAccessVerifier) Verify(token string) (string, error) {
	parsed, err := jwt.Parse(token, v.jwks.Keyfunc)
	if err != nil {
		return "", fmt.Errorf("parse token: %w", err)
	}
	if !parsed.Valid {
		return "", errors.New("token invalid")
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return "", errors.New("unexpected claims type")
	}

	if !claims.VerifyIssuer(v.issuer, true) {
		return "", errors.New("issuer mismatch")
	}
	matched := false
	for _, aud := range v.audiences {
		if claims.VerifyAudience(aud, true) {
			matched = true
			break
		}
	}
	if !matched {
		return "", fmt.Errorf("audience mismatch: expected one of %v, got %v", v.audiences, claims["aud"])
	}
	if !claims.VerifyExpiresAt(time.Now().Unix(), true) {
		return "", errors.New("token expired")
	}

	email, _ := claims["email"].(string)
	if email == "" {
		return "", errors.New("missing email claim")
	}
	return email, nil
}
