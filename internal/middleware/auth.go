package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/LMV0911/Go_e-commerce/internal/auth"
)

type contextKey string

const (
	ClaimsKey contextKey = "claims"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				tokenStr := parts[1]
				claims, err := auth.ValidateToken(tokenStr)
				if err == nil && claims != nil {
					ctx := context.WithValue(r.Context(), ClaimsKey, claims)
					r = r.WithContext(ctx)
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}

func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetClaims(r.Context())
		if !ok || claims == nil {
			authHeader := r.Header.Get("Authorization")
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
				tokenClaims, err := auth.ValidateToken(parts[1])
				if err == nil && tokenClaims != nil {
					ctx := context.WithValue(r.Context(), ClaimsKey, tokenClaims)
					next(w, r.WithContext(ctx))
					return
				}
			}

			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Authentification requise : token JWT manquant ou invalide",
			})
			return
		}
		next(w, r)
	}
}

func RequireRole(role string, next http.HandlerFunc) http.HandlerFunc {
	return RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := GetClaims(r.Context())
		if !ok || claims.Role != role {
			w.Header().Set("Content-Type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Accès refusé : privilèges insuffisants",
			})
			return
		}
		next(w, r)
	})
}

func GetClaims(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value(ClaimsKey).(*auth.Claims)
	return claims, ok
}

func GetUserID(ctx context.Context) (int, bool) {
	claims, ok := GetClaims(ctx)
	if !ok || claims == nil {
		return 0, false
	}
	return claims.UserID, true
}

func GetUserRole(ctx context.Context) (string, bool) {
	claims, ok := GetClaims(ctx)
	if !ok || claims == nil {
		return "", false
	}
	return claims.Role, true
}

func GetUserEmail(ctx context.Context) (string, bool) {
	claims, ok := GetClaims(ctx)
	if !ok || claims == nil {
		return "", false
	}
	return claims.Email, true
}
