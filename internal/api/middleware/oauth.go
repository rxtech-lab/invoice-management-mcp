package middleware

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/rxtech-lab/invoice-management/internal/utils"
)

// OAuthConfig holds configuration for OAuth authentication
type OAuthConfig struct {
	// JWKSEndpoint is the URL to fetch JWKS from (e.g., "https://auth.example.com/.well-known/jwks.json")
	JWKSEndpoint string

	// Issuer to validate (optional, if empty, issuer is not validated)
	Issuer string

	// Audience to validate (optional)
	Audience string

	// RefreshInterval for JWKS key refresh (default: 1 hour)
	RefreshInterval time.Duration

	// RefreshTimeout for JWKS fetch timeout (default: 10 seconds)
	RefreshTimeout time.Duration
}

// OAuthAuthenticator handles JWT validation against JWKS endpoint
type OAuthAuthenticator struct {
	jwks   keyfunc.Keyfunc
	config OAuthConfig
}

// NewOAuthAuthenticator creates a new OAuth authenticator
func NewOAuthAuthenticator(config OAuthConfig) (*OAuthAuthenticator, error) {
	// Set defaults
	if config.RefreshInterval == 0 {
		config.RefreshInterval = time.Hour
	}
	if config.RefreshTimeout == 0 {
		config.RefreshTimeout = 10 * time.Second
	}

	// Create JWKS keyfunc with automatic refresh
	ctx, cancel := context.WithTimeout(context.Background(), config.RefreshTimeout)
	defer cancel()

	k, err := keyfunc.NewDefaultCtx(ctx, []string{config.JWKSEndpoint})
	if err != nil {
		return nil, fmt.Errorf("failed to create JWKS keyfunc: %w", err)
	}

	return &OAuthAuthenticator{
		jwks:   k,
		config: config,
	}, nil
}

// ValidateToken validates a JWT token and returns the claims
func (a *OAuthAuthenticator) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	// Parse and validate the token
	token, err := jwt.Parse(tokenString, a.jwks.Keyfunc,
		jwt.WithValidMethods([]string{"RS256", "RS384", "RS512", "ES256", "ES384", "ES512"}),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Validate issuer if configured
	if a.config.Issuer != "" {
		iss, _ := claims.GetIssuer()
		if iss != a.config.Issuer {
			return nil, fmt.Errorf("invalid issuer: expected %s, got %s", a.config.Issuer, iss)
		}
	}

	// Validate audience if configured
	if a.config.Audience != "" {
		aud, _ := claims.GetAudience()
		found := false
		for _, audValue := range aud {
			if audValue == a.config.Audience {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("invalid audience")
		}
	}

	return claims, nil
}

// ExtractUser extracts AuthenticatedUser from JWT claims
func (a *OAuthAuthenticator) ExtractUser(claims jwt.MapClaims) *utils.AuthenticatedUser {
	user := &utils.AuthenticatedUser{}

	// Extract subject (user ID)
	if sub, ok := claims["sub"].(string); ok {
		user.Sub = sub
	}

	// Extract roles (common claim names: "roles", "role", "realm_access.roles")
	user.Roles = extractStringSlice(claims, "roles")
	if len(user.Roles) == 0 {
		user.Roles = extractStringSlice(claims, "role")
	}

	// Extract scopes (common claim names: "scope", "scp")
	if scope, ok := claims["scope"].(string); ok {
		user.Scopes = strings.Split(scope, " ")
	} else {
		user.Scopes = extractStringSlice(claims, "scp")
	}

	return user
}

// Close is a no-op for cleanup compatibility
// keyfunc v3 handles lifecycle automatically with internal goroutines
func (a *OAuthAuthenticator) Close() {
	// No explicit cleanup needed in keyfunc v3
}

// extractStringSlice extracts a string slice from claims
func extractStringSlice(claims jwt.MapClaims, key string) []string {
	if val, ok := claims[key]; ok {
		switch v := val.(type) {
		case []interface{}:
			result := make([]string, 0, len(v))
			for _, item := range v {
				if str, ok := item.(string); ok {
					result = append(result, str)
				}
			}
			return result
		case []string:
			return v
		}
	}
	return nil
}

// FiberOAuthMiddleware creates a Fiber middleware for OAuth JWT authentication
func FiberOAuthMiddleware(authenticator *OAuthAuthenticator, onSuccess func(*fiber.Ctx, *utils.AuthenticatedUser) error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return c.Next() // No auth header, let other middleware handle it
		}

		// Check for Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return c.Next() // Not a Bearer token, let other middleware handle it
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate token
		claims, err := authenticator.ValidateToken(tokenString)
		if err != nil {
			log.Printf("OAuth token validation failed: %v", err)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		// Extract user from claims
		user := authenticator.ExtractUser(claims)

		// Call success callback
		if onSuccess != nil {
			if err := onSuccess(c, user); err != nil {
				return err
			}
		}

		return c.Next()
	}
}
