package middleware

import (
	"context"
	"net/http"
	"strings"

	"skillhub/backend/internal/auth"
)

type contextKey string

const actorContextKey contextKey = "authActor"

type TokenAuthenticator interface {
	AuthenticateToken(token string) (auth.Actor, error)
}

func RequireAuth(authenticator TokenAuthenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				http.Error(w, `{"error":"missing bearer token"}`, http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(header, "Bearer ")
			actor, err := authenticator.AuthenticateToken(token)
			if err != nil {
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), actorContextKey, actor)))
		})
	}
}

// bearerOrQueryToken returns a JWT from Authorization: Bearer or from the access_token query param
// (EventSource cannot send custom headers, so the web app passes the token in the query string).
func bearerOrQueryToken(r *http.Request) string {
	header := r.Header.Get("Authorization")
	if strings.HasPrefix(header, "Bearer ") {
		return strings.TrimPrefix(header, "Bearer ")
	}
	return r.URL.Query().Get("access_token")
}

// RequireAuthFlexible is like RequireAuth but also accepts access_token in the query string.
func RequireAuthFlexible(authenticator TokenAuthenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerOrQueryToken(r)
			if token == "" {
				http.Error(w, `{"error":"missing bearer token"}`, http.StatusUnauthorized)
				return
			}

			actor, err := authenticator.AuthenticateToken(token)
			if err != nil {
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), actorContextKey, actor)))
		})
	}
}

func ActorFromContext(ctx context.Context) (auth.Actor, bool) {
	actor, ok := ctx.Value(actorContextKey).(auth.Actor)
	return actor, ok
}
