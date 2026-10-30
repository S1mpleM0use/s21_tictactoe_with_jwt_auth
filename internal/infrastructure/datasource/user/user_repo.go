package user

import (
	"context"
	"errors"
	"fmt"

	domainModel "tictactoe/internal/domain/model"
	"tictactoe/internal/infrastructure/datasource/mapper"
	db "tictactoe/internal/infrastructure/datasource/storage/pgsql"

	"github.com/google/uuid"
)

type userRepository struct {
	storage *db.GameStorage
}

func NewUserRepository(storage *db.GameStorage) UserRepository {
	return &userRepository{storage: storage}
}

func (r *userRepository) SaveUser(ctx context.Context, user domainModel.User) error {
	dto := mapper.DomainUserToInfraUser(user)
	if err := r.storage.SaveUser(ctx, dto); err != nil {
		if errors.Is(err, db.ErrLoginAlreadyTaken) {
			return fmt.Errorf("%w: %s", ErrLoginAlreadyTaken, user.Login)
		}
		return fmt.Errorf("save user: %w", err)
	}
	return nil
}

func (r *userRepository) GetUserByLogin(ctx context.Context, login string) (domainModel.User, error) {
	foundUser, err := r.storage.GetUserByLogin(ctx, login)
	if errors.Is(err, db.ErrUserNotFound) {
		return domainModel.User{}, fmt.Errorf("%w: login %q", ErrUserNotFound, login)
	}
	if err != nil {
		return domainModel.User{}, fmt.Errorf("get user by login: %w", err)
	}
	return mapper.InfraUserToDomainUser(foundUser), nil
}

func (r *userRepository) GetUserByID(ctx context.Context, id uuid.UUID) (domainModel.User, error) {
	foundUser, err := r.storage.GetUserByID(ctx, id)
	if errors.Is(err, db.ErrUserNotFound) {
		return domainModel.User{}, fmt.Errorf("%w: id %s", ErrUserNotFound, id)
	}
	if err != nil {
		return domainModel.User{}, fmt.Errorf("get user by id: %w", err)
	}
	return mapper.InfraUserToDomainUser(foundUser), nil
}
