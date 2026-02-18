package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/aiox-platform/aiox/internal/api"
)

type contextKey string

const UserClaimsKey contextKey = "user_claims"

func Middleware(svc *Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				api.HandleError(w, api.ErrUnauthorized)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
				api.HandleError(w, api.ErrUnauthorized)
				return
			}

			claims, err := svc.jwt.ValidateAccessToken(parts[1])
			if err != nil {
				api.HandleError(w, api.ErrInvalidToken)
				return
			}

			ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserClaims(ctx context.Context) *AccessClaims {
	claims, _ := ctx.Value(UserClaimsKey).(*AccessClaims)
	return claims
}
