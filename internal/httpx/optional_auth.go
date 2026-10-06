package httpx

import (
	"context"
	"net/http"
	"strings"
)

func OptionalAuth(parser AuthParser, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if strings.HasPrefix(header, "Bearer ") {
			raw := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
			if claims, err := parser.ParseAccess(raw); err == nil {
				ctx := context.WithValue(r.Context(), ctxUserID, claims.UserID)
				ctx = context.WithValue(ctx, ctxUserRole, claims.Role)
				r = r.WithContext(ctx)
			}
		}
		next.ServeHTTP(w, r)
	})
}
