package handler

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"tictactoe/internal/app"

	"github.com/google/uuid"
)

type contextKey string

const userUUIDKey contextKey = "userUUID"

func userUUID(ctx context.Context) (uuid.UUID, error) {
	id, ok := ctx.Value(userUUIDKey).(uuid.UUID)
	if !ok || id == uuid.Nil {
		return uuid.Nil, errors.New("no authenticated user in context")
	}
	return id, nil
}

type UserAuthenticator struct {
	jwtProv app.JWTProvider
}

func NewUserAuthenticator(jwtProv app.JWTProvider) *UserAuthenticator {
	return &UserAuthenticator{jwtProv: jwtProv}
}

func (a *UserAuthenticator) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := extractBearerToken(r.Header.Get("Authorization"))
		if err != nil {
			if isAuthenticationError(err) {
				w.Header().Set("WWW-Authenticate", `Bearer realm="tic-tac-toe"`)
				writeJSONError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			log.Printf("authentication service error: %v", err)
			writeJSONError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		if err := a.jwtProv.ValidateAccessToken(token); err != nil {
			w.Header().Set("WWW-Authenticate", `Bearer realm="tic-tac-toe"`)
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		userID, err := a.jwtProv.ExtractUUIDfromToken(token)
		if err != nil {
			w.Header().Set("WWW-Authenticate", `Bearer realm="tic-tac-toe"`)
			writeJSONError(w, http.StatusUnauthorized, "unauthorized")
			return
		}

		ctx := context.WithValue(r.Context(), userUUIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})

}

func extractBearerToken(header string) (string, error) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", app.ErrAuthNotBearer
	}

	token := strings.TrimSpace(strings.TrimPrefix(header, prefix))
	if token == "" {
		return "", app.ErrInvalidToken
	}

	return token, nil
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(map[string]string{"error": message}); err != nil {
		log.Printf("encode error response: %v", err)
	}
}
