package pgsql

import "errors"

var (
	ErrGameNotFound      = errors.New("game not found")
	ErrUserNotFound      = errors.New("user not found")
	ErrLoginAlreadyTaken = errors.New("login already taken")
)
