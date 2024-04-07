package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rovilay/auth-service/utils"
)

type contextKey string

var userIDKey contextKey = "userID"

func (h *UserHandler) MiddlewareAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization") // Assuming token in "Authorization" header
		if tokenString == "" {
			ErrUnauthorized(w, utils.ErrMissingAuthToken)
			return
		}

		userID, err := utils.ValidateJWT(tokenString)
		if err != nil {
			ErrUnauthorized(w, err)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ErrUnauthorized is a helper for consistent unauthorized responses
func ErrUnauthorized(w http.ResponseWriter, err error) {
	errRes := fmt.Sprintf(`{"error": "%v"}`, err.Error())
	http.Error(w, errRes, http.StatusUnauthorized)
}
