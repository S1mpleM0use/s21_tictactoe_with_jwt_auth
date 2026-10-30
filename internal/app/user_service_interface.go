package app

import (
	"context"
	domainModel "tictactoe/internal/domain/model"

	"github.com/google/uuid"
)

type UserService interface {
	Create(ctx context.Context, req domainModel.SignUpRequest) (domainModel.User, error)
	VerifyCredentials(ctx context.Context, login, password string) (domainModel.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (domainModel.User, error)
}
