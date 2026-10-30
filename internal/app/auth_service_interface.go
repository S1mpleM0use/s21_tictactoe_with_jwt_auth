package app

import (
	"context"
	domainModel "tictactoe/internal/domain/model"
)

type AuthService interface {
	Register(ctx context.Context, req domainModel.SignUpRequest) (domainModel.User, error)
	Authenticate(ctx context.Context, jwtReq domainModel.JwtRequest) (domainModel.JwtResponse, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (domainModel.JwtResponse, error)
	RefreshRefreshToken(ctx context.Context, refreshToken string) (domainModel.JwtResponse, error)
}
