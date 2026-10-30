package app

import (
	"context"
	"errors"
	"fmt"
	"strings"

	domainModel "tictactoe/internal/domain/model"
)

const tokenTypeBearer = "Bearer"

type authService struct {
	userService UserService
	jwtProv     JWTProvider
}

func NewAuthService(userService UserService, jwtProv JWTProvider) AuthService {
	return &authService{
		userService: userService,
		jwtProv:     jwtProv,
	}
}

func (s *authService) Register(ctx context.Context, req domainModel.SignUpRequest) (domainModel.User, error) {
	return s.userService.Create(ctx, req)
}

func (s *authService) Authenticate(ctx context.Context, jwtReq domainModel.JwtRequest) (domainModel.JwtResponse, error) {
	login := strings.TrimSpace(jwtReq.Login)
	if login == "" || jwtReq.Password == "" {
		return domainModel.JwtResponse{}, ErrInvalidCredentials
	}

	user, err := s.userService.VerifyCredentials(ctx, login, jwtReq.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return domainModel.JwtResponse{}, ErrInvalidCredentials
		}
		return domainModel.JwtResponse{}, fmt.Errorf("verify credentials: %w", err)
	}

	accessToken, err := s.jwtProv.GenerateAccessToken(user)
	if err != nil {
		return domainModel.JwtResponse{}, fmt.Errorf("generate access token: %w", err)
	}

	refreshToken, err := s.jwtProv.GenerateRefreshToken(user)
	if err != nil {
		return domainModel.JwtResponse{}, fmt.Errorf("generate refresh token: %w", err)
	}

	return domainModel.JwtResponse{
		Type:         tokenTypeBearer,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *authService) RefreshAccessToken(ctx context.Context, refreshToken string) (domainModel.JwtResponse, error) {
	user, err := s.validateRefreshAndLoadUser(ctx, refreshToken)
	if err != nil {
		return domainModel.JwtResponse{}, err
	}

	accessToken, err := s.jwtProv.GenerateAccessToken(user)
	if err != nil {
		return domainModel.JwtResponse{}, fmt.Errorf("generate access token: %w", err)
	}

	return domainModel.JwtResponse{
		Type:         tokenTypeBearer,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *authService) RefreshRefreshToken(ctx context.Context, refreshToken string) (domainModel.JwtResponse, error) {
	user, err := s.validateRefreshAndLoadUser(ctx, refreshToken)
	if err != nil {
		return domainModel.JwtResponse{}, err
	}

	newAccessToken, err := s.jwtProv.GenerateAccessToken(user)
	if err != nil {
		return domainModel.JwtResponse{}, fmt.Errorf("generate access token: %w", err)
	}

	newRefreshToken, err := s.jwtProv.GenerateRefreshToken(user)
	if err != nil {
		return domainModel.JwtResponse{}, fmt.Errorf("generate refresh token: %w", err)
	}

	return domainModel.JwtResponse{
		Type:         tokenTypeBearer,
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *authService) validateRefreshAndLoadUser(ctx context.Context, refreshToken string) (domainModel.User, error) {
	refreshToken = strings.TrimSpace(refreshToken)
	if refreshToken == "" {
		return domainModel.User{}, ErrInvalidToken
	}

	if err := s.jwtProv.ValidateRefreshToken(refreshToken); err != nil {
		return domainModel.User{}, ErrInvalidToken
	}

	userID, err := s.jwtProv.ExtractUUIDfromToken(refreshToken)
	if err != nil {
		return domainModel.User{}, ErrInvalidToken
	}

	user, err := s.userService.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return domainModel.User{}, ErrInvalidToken
		}
		return domainModel.User{}, fmt.Errorf("get user for token refresh: %w", err)
	}

	return user, nil
}
