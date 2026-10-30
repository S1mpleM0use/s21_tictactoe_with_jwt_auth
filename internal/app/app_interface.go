package app

import (
	"context"

	domainModel "tictactoe/internal/domain/model"

	"github.com/google/uuid"
)

type AppService interface {
	CreateGame(ctx context.Context, hostID uuid.UUID, mode string) (domainModel.Game, error)
	JoinGame(ctx context.Context, gameID, userID uuid.UUID) (domainModel.Game, error)
	ApplyMove(ctx context.Context, gameID, userID uuid.UUID, row, col int) (domainModel.Game, error)
	GetWaitingGames(ctx context.Context) ([]domainModel.Game, error)
	GetFinishedGamesByUserID(ctx context.Context, userID uuid.UUID) ([]domainModel.Game, error)
	GetCurrentGame(ctx context.Context, userID uuid.UUID) (domainModel.Game, error)
	GetLeaderboard(ctx context.Context, limit int) ([]domainModel.LeaderboardEntry, error)
}

type Transactor interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
