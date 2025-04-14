package authmw

import (
	"AvitoPVZ/internal/config"
	"AvitoPVZ/internal/user"
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt"
)

const userClaims = "UserClaims"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "No Authorization header", http.StatusUnauthorized)
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid Authorization header", http.StatusUnauthorized)
			return
		}
		tokenStr := parts[1]

		claims := &user.Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv(config.JwtSecret)), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userClaims, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetClaims(r *http.Request) *user.Claims {
	claims, ok := r.Context().Value(userClaims).(*user.Claims)
	if !ok {
		return nil
	}
	return claims
}
