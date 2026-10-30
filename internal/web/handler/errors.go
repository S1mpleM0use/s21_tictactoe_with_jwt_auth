package handler

import (
	"errors"
	"tictactoe/internal/app"
)

func isAuthenticationError(err error) bool {
	return errors.Is(err, app.ErrInvalidCredentials) ||
		errors.Is(err, app.ErrInvalidToken) ||
		errors.Is(err, app.ErrAuthNotBearer)
}
