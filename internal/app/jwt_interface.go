package app

import (
	domainModel "tictactoe/internal/domain/model"

	"github.com/google/uuid"
)

type JWTProvider interface {
	GenerateAccessToken(user domainModel.User) (string, error)
	GenerateRefreshToken(user domainModel.User) (string, error)
	ValidateAccessToken(token string) error
	ValidateRefreshToken(token string) error
	ExtractUUIDfromToken(token string) (uuid.UUID, error)
}
