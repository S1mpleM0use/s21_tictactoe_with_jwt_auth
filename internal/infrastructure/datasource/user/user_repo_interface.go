package user

import (
	"context"
	domainModel "tictactoe/internal/domain/model"

	"github.com/google/uuid"
)

type UserRepository interface {
	SaveUser(ctx context.Context, user domainModel.User) error
	GetUserByLogin(ctx context.Context, login string) (domainModel.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (domainModel.User, error)
}
