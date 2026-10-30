package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	domainModel "tictactoe/internal/domain/model"
	userRepo "tictactoe/internal/infrastructure/datasource/user"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	userRepo userRepo.UserRepository
}

func NewUserService(repo userRepo.UserRepository) UserService {
	return &userService{userRepo: repo}
}

func (u *userService) Create(ctx context.Context, req domainModel.SignUpRequest) (domainModel.User, error) {
	req.Login = strings.TrimSpace(req.Login)
	if req.Login == "" {
		return domainModel.User{}, ErrLoginRequired
	}
	if utf8.RuneCountInString(req.Login) > 20 {
		return domainModel.User{}, ErrLoginTooLong
	}
	if len(req.Password) < 4 {
		return domainModel.User{}, ErrPasswordTooShort
	}

	_, err := u.userRepo.GetUserByLogin(ctx, req.Login)
	switch {
	case err == nil:
		return domainModel.User{}, ErrLoginAlreadyTaken
	case errors.Is(err, userRepo.ErrUserNotFound):
		// Логин свободен
	default:
		return domainModel.User{}, fmt.Errorf("check login availability: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return domainModel.User{}, fmt.Errorf("hash password: %w", err)
	}

	newUser := domainModel.User{
		ID:       uuid.New(),
		Login:    req.Login,
		Password: string(hash),
	}

	if err := u.userRepo.SaveUser(ctx, newUser); err != nil {
		if errors.Is(err, userRepo.ErrLoginAlreadyTaken) {
			return domainModel.User{}, ErrLoginAlreadyTaken
		}
		return domainModel.User{}, fmt.Errorf("save user: %w", err)
	}

	return newUser, nil
}

func (u *userService) VerifyCredentials(ctx context.Context, login, password string) (domainModel.User, error) {
	foundUser, err := u.userRepo.GetUserByLogin(ctx, login)
	if errors.Is(err, userRepo.ErrUserNotFound) {
		return domainModel.User{}, ErrInvalidCredentials
	}
	if err != nil {
		return domainModel.User{}, fmt.Errorf("get user by login: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(password)); err != nil {
		return domainModel.User{}, ErrInvalidCredentials
	}

	return foundUser, nil
}

func (u *userService) GetByID(ctx context.Context, id uuid.UUID) (domainModel.User, error) {
	foundUser, err := u.userRepo.GetUserByID(ctx, id)
	if errors.Is(err, userRepo.ErrUserNotFound) {
		return domainModel.User{}, fmt.Errorf("%w: %s", ErrUserNotFound, id)
	}
	if err != nil {
		return domainModel.User{}, fmt.Errorf("get user by id: %w", err)
	}
	return foundUser, nil
}
