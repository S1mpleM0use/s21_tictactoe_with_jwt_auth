package storage

import (
	"context"
	"errors"
	"fmt"

	domainModel "tictactoe/internal/domain/model"
	"tictactoe/internal/infrastructure/datasource/mapper"
	db "tictactoe/internal/infrastructure/datasource/storage/pgsql"

	"github.com/google/uuid"
)

type Repository struct {
	storage *db.GameStorage
}

func NewRepository(storage *db.GameStorage) GameRepository {
	return &Repository{storage: storage}
}

func (r *Repository) SaveGame(ctx context.Context, game domainModel.Game) error {
	dbModel := mapper.ToModel(game)
	if err := r.storage.StoreGame(ctx, dbModel); err != nil {
		return fmt.Errorf("save game %s: %w", game.ID, err)
	}
	return nil
}

func (r *Repository) GetGameByID(ctx context.Context, id uuid.UUID) (domainModel.Game, error) {
	game, err := r.storage.LoadGameByID(ctx, id)
	if errors.Is(err, db.ErrGameNotFound) {
		return domainModel.Game{}, fmt.Errorf("%w: %s", ErrGameNotFound, id)
	}
	if err != nil {
		return domainModel.Game{}, fmt.Errorf("load game %s: %w", id, err)
	}
	return mapper.ToDomain(game), nil
}

func (r *Repository) GetGamesByState(ctx context.Context, state string) ([]domainModel.Game, error) {
	models, err := r.storage.LoadGamesByState(ctx, state)
	if err != nil {
		return nil, fmt.Errorf("load games by state %q: %w", state, err)
	}

	games := make([]domainModel.Game, 0, len(models))
	for _, model := range models {
		games = append(games, mapper.ToDomain(model))
	}
	return games, nil
}

func (r *Repository) GetCompletedGamesByPlayerID(ctx context.Context, playerID uuid.UUID) ([]domainModel.Game, error) {
	models, err := r.storage.LoadCompletedGamesByUser(ctx, playerID)
	if err != nil {
		return nil, fmt.Errorf("load completed games by user %s: %w", playerID, err)
	}

	games := make([]domainModel.Game, 0, len(models))
	for _, model := range models {
		games = append(games, mapper.ToDomain(model))
	}
	return games, nil
}

func (r *Repository) GetCurrentGame(ctx context.Context, userID uuid.UUID) (domainModel.Game, error) {
	game, err := r.storage.LoadCurrentGame(ctx, userID)
	if errors.Is(err, db.ErrGameNotFound) {
		return domainModel.Game{}, fmt.Errorf("%w for user %s", ErrGameNotFound, userID)
	}
	if err != nil {
		return domainModel.Game{}, fmt.Errorf("load current game for user %s: %w", userID, err)
	}
	return mapper.ToDomain(game), nil
}

func (r *Repository) GetLeaderboard(ctx context.Context, limit int) ([]domainModel.LeaderboardEntry, error) {
	dtos, err := r.storage.LoadLeaderboard(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("load leaderboard: %w", err)
	}

	entries := make([]domainModel.LeaderboardEntry, 0, len(dtos))
	for _, dto := range dtos {
		entries = append(entries, mapper.InfraLeaderboardToDomain(dto))
	}
	return entries, nil
}
