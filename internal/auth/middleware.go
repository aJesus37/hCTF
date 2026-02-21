package auth

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	UserContextKey contextKey = "user"
)

type Claims struct {
	UserID  string `json:"user_id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	IsAdmin bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

var jwtSecret []byte

// SetJWTSecret sets the JWT signing secret. Must be called before server starts.
func SetJWTSecret(secret string) error {
	if len(secret) < 32 {
		return errors.New("JWT secret must be at least 32 characters")
	}
	jwtSecret = []byte(secret)
	return nil
}

// GetJWTSecret returns the current JWT secret (used internally)
func GetJWTSecret() []byte {
	return jwtSecret
}

// GenerateToken creates a JWT token for the user
func GenerateToken(userID, email, name string, isAdmin bool) (string, error) {
	if len(jwtSecret) == 0 {
		return "", errors.New("JWT secret not configured")
	}
	claims := &Claims{
		UserID:  userID,
		Email:   email,
		Name:    name,
		IsAdmin: isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(168 * time.Hour)), // 7 days
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// ValidateToken verifies and parses a JWT token
func ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

// AuthMiddleware validates JWT from cookie or Authorization header
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenString string

		// Try cookie first
		cookie, err := r.Cookie("auth_token")
		if err == nil {
			tokenString = cookie.Value
		} else {
			// Try Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		if tokenString != "" {
			claims, err := ValidateToken(tokenString)
			if err == nil {
				ctx := context.WithValue(r.Context(), UserContextKey, claims)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// RequireAuth ensures the user is authenticated
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := GetUserFromContext(r.Context())
		if claims == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireAdmin ensures the user is an admin
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := GetUserFromContext(r.Context())
		if claims == nil || !claims.IsAdmin {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// GetUserFromContext retrieves user claims from context
func GetUserFromContext(ctx context.Context) *Claims {
	claims, ok := ctx.Value(UserContextKey).(*Claims)
	if !ok {
		return nil
	}
	return claims
}
