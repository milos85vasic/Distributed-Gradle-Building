package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"distributed-gradle-building/errors"
	"github.com/dgrijalva/jwt-go"
)

// AuthService handles authentication and authorization
type AuthService struct {
	secretKey     []byte
	tokenTTL      time.Duration
	allowedTokens map[string]bool
	adminTokens   map[string]bool
}

// Claims represents JWT claims
type Claims struct {
	UserID      string   `json:"user_id"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
	jwt.StandardClaims
}

// NewAuthService creates a new authentication service
func NewAuthService(secretKey string, tokenTTL time.Duration) *AuthService {
	return &AuthService{
		secretKey:     []byte(secretKey),
		tokenTTL:      tokenTTL,
		allowedTokens: make(map[string]bool),
		adminTokens:   make(map[string]bool),
	}
}

// GenerateToken generates a new JWT token
func (a *AuthService) GenerateToken(userID, role string, permissions []string) (string, error) {
	claims := &Claims{
		UserID:      userID,
		Role:        role,
		Permissions: permissions,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(a.tokenTTL).Unix(),
			IssuedAt:  time.Now().Unix(),
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(a.secretKey)
}

// ValidateToken validates a JWT token
func (a *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.secretKey, nil
	})

	if err != nil {
		return nil, errors.NewAPIError(errors.ErrCodeUnauthorized, "Invalid token")
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.NewAPIError(errors.ErrCodeUnauthorized, "Invalid token claims")
}

// AddAllowedToken adds a token to the allowed list
func (a *AuthService) AddAllowedToken(token string) {
	a.allowedTokens[token] = true
}

// AddAdminToken adds an admin token
func (a *AuthService) AddAdminToken(token string) {
	a.adminTokens[token] = true
}

// IsAdmin checks if a token has admin privileges
func (a *AuthService) IsAdmin(token string) bool {
	return a.adminTokens[token]
}

// IsAllowed checks if a token is allowed
func (a *AuthService) IsAllowed(token string) bool {
	return a.allowedTokens[token]
}

// AuthMiddleware creates authentication middleware
func (a *AuthService) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip authentication for health check
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			sendAuthError(w, errors.NewAPIError(errors.ErrCodeUnauthorized, "Missing authorization header"))
			return
		}

		// Extract token from "Bearer <token>"
		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			sendAuthError(w, errors.NewAPIError(errors.ErrCodeUnauthorized, "Invalid authorization header format"))
			return
		}

		token := tokenParts[1]

		// Check static token list first (for service tokens)
		if a.IsAllowed(token) {
			next.ServeHTTP(w, r)
			return
		}

		// Validate JWT token
		claims, err := a.ValidateToken(token)
		if err != nil {
			sendAuthError(w, err)
			return
		}

		// Add claims to context
		ctx := context.WithValue(r.Context(), "claims", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// AdminMiddleware creates admin-only middleware
func (a *AuthService) AdminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			sendAuthError(w, errors.NewAPIError(errors.ErrCodeUnauthorized, "Missing authorization header"))
			return
		}

		tokenParts := strings.SplitN(authHeader, " ", 2)
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			sendAuthError(w, errors.NewAPIError(errors.ErrCodeUnauthorized, "Invalid authorization header format"))
			return
		}

		token := tokenParts[1]

		// Check admin privileges
		if !a.IsAdmin(token) {
			// Check JWT claims for admin role
			claims, err := a.ValidateToken(token)
			if err != nil || claims.Role != "admin" {
				sendAuthError(w, errors.NewAPIError(errors.ErrCodeForbidden, "Admin access required"))
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// CORSMiddleware creates CORS middleware
func CORSMiddleware(allowedOrigins, allowedMethods, allowedHeaders []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				if allowedOrigin == "*" || allowedOrigin == origin {
					allowed = true
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
			}

			// Set other CORS headers
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "3600")

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// sendAuthError sends authentication error response
func sendAuthError(w http.ResponseWriter, err error) {
	if apiErr, ok := err.(*errors.APIError); ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(apiErr.HTTPStatus)
		json.NewEncoder(w).Encode(apiErr)
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"error": "Unauthorized"})
	}
}

// GetClaimsFromContext extracts claims from request context
func GetClaimsFromContext(r *http.Request) (*Claims, bool) {
	claims, ok := r.Context().Value("claims").(*Claims)
	return claims, ok
}

// HasPermission checks if user has specific permission
func HasPermission(claims *Claims, permission string) bool {
	for _, p := range claims.Permissions {
		if p == permission || p == "*" {
			return true
		}
	}
	return false
}

// RequirePermission middleware to check specific permission
func (a *AuthService) RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := GetClaimsFromContext(r)
			if !ok {
				sendAuthError(w, errors.NewAPIError(errors.ErrCodeUnauthorized, "User not authenticated"))
				return
			}

			if !HasPermission(claims, permission) {
				sendAuthError(w, errors.NewAPIError(errors.ErrCodeForbidden, fmt.Sprintf("Permission '%s' required", permission)))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
