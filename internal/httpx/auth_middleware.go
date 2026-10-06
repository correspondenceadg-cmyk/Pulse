package httpx

import (
	"context"
	"net/http"
	"strings"
)

type AuthParser interface {
	ParseAccess(raw string) (*AuthClaims, error)
}

type AuthClaims struct {
	UserID string
	Role   string
}

type authCtxKey int

const (
	ctxUserID authCtxKey = iota
	ctxUserRole
)

func UserIDFrom(ctx context.Context) string {
	v, _ := ctx.Value(ctxUserID).(string)
	return v
}

func UserRoleFrom(ctx context.Context) string {
	v, _ := ctx.Value(ctxUserRole).(string)
	return v
}

func RequireAuth(parser AuthParser, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			Error(w, r, http.StatusUnauthorized, "missing bearer token")
			return
		}
		raw := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		claims, err := parser.ParseAccess(raw)
		if err != nil {
			Error(w, r, http.StatusUnauthorized, "invalid or expired token")
			return
		}
		ctx := context.WithValue(r.Context(), ctxUserID, claims.UserID)
		ctx = context.WithValue(ctx, ctxUserRole, claims.Role)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}