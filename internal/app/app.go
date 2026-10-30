package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	domainModel "tictactoe/internal/domain/model"
	"tictactoe/internal/domain/service"
	"tictactoe/internal/infrastructure/datasource/storage"

	"github.com/google/uuid"
)

type appService struct {
	gameService service.GameService
	gameRepo    storage.GameRepository
	txManager   Transactor
}

func NewAppService(gameService service.GameService, gameRepo storage.GameRepository, txMan Transactor) AppService {
	return &appService{
		gameService: gameService,
		gameRepo:    gameRepo,
		txManager:   txMan,
	}
}

func (a *appService) CreateGame(ctx context.Context, hostID uuid.UUID, mode string) (domainModel.Game, error) {
	if mode != domainModel.ModeVSPlayer && mode != domainModel.ModeVSComputer {
		return domainModel.Game{}, ErrInvalidMode
	}

	game := domainModel.Game{
		ID:        uuid.New(),
		Board:     createEmptyBoard(domainModel.BoardDimension),
		GameState: domainModel.StateWaiting,
		Mode:      mode,
		HostID:    hostID,
		Symbols: map[string]string{
			hostID.String(): domainModel.SymbolX,
		},
		CreatedAt: time.Now().UTC(),
	}

	if mode == domainModel.ModeVSComputer {
		compID := domainModel.ComputerID
		nextID := hostID
		game.GameState = domainModel.StatePlaying
		game.OpponentID = &compID
		game.NextPlayerID = &nextID
		game.Symbols[compID.String()] = domainModel.SymbolO
	}

	if err := a.gameRepo.SaveGame(ctx, game); err != nil {
		return domainModel.Game{}, fmt.Errorf("save new game: %w", err)
	}
	return game, nil
}

func (a *appService) JoinGame(ctx context.Context, gameID, userID uuid.UUID) (domainModel.Game, error) {
	var resultGame domainModel.Game

	err := a.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		game, err := a.gameRepo.GetGameByID(txCtx, gameID)
		if err != nil {
			if errors.Is(err, storage.ErrGameNotFound) {
				return ErrGameNotFound
			}
			return fmt.Errorf("load game before joining: %w", err)
		}

		if game.Mode != domainModel.ModeVSPlayer {
			return ErrNotVSPlayer
		}
		if game.GameState != domainModel.StateWaiting || game.OpponentID != nil {
			return ErrNotWaiting
		}
		if userID == game.HostID {
			return ErrCreatorCannotJoin
		}

		if game.Symbols == nil {
			game.Symbols = make(map[string]string, 2)
		}
		oppID := userID
		hostID := game.HostID
		game.OpponentID = &oppID
		game.Symbols[hostID.String()] = domainModel.SymbolX
		game.Symbols[oppID.String()] = domainModel.SymbolO
		game.GameState = domainModel.StatePlaying
		game.NextPlayerID = &hostID

		if err := a.gameRepo.SaveGame(txCtx, game); err != nil {
			return fmt.Errorf("save joined game: %w", err)
		}
		resultGame = game
		return nil
	})

	return resultGame, err
}

func (a *appService) ApplyMove(ctx context.Context, gameID, userID uuid.UUID, row, col int) (domainModel.Game, error) {
	var resultGame domainModel.Game

	err := a.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		game, err := a.gameRepo.GetGameByID(txCtx, gameID)
		if err != nil {
			if errors.Is(err, storage.ErrGameNotFound) {
				return ErrGameNotFound
			}
			return fmt.Errorf("load game before move: %w", err)
		}

		if game.GameState != domainModel.StatePlaying {
			return ErrGameNotPlaying
		}
		if userID != game.HostID && (game.OpponentID == nil || userID != *game.OpponentID) {
			return ErrNotYourGame
		}
		if game.NextPlayerID == nil || userID != *game.NextPlayerID {
			return ErrNotYourTurn
		}
		if row < 0 || row >= domainModel.BoardDimension || col < 0 || col >= domainModel.BoardDimension {
			return ErrCellOutOfBounds
		}
		if len(game.Board) != domainModel.BoardDimension {
			return fmt.Errorf("invalid board: expected %d rows", domainModel.BoardDimension)
		}
		for i := range game.Board {
			if len(game.Board[i]) != domainModel.BoardDimension {
				return fmt.Errorf("invalid board: row %d has %d columns", i, len(game.Board[i]))
			}
		}
		if game.Board[row][col] != domainModel.Empty {
			return ErrCellOccupied
		}

		symbol, ok := game.Symbols[userID.String()]
		if !ok {
			return ErrNotYourGame
		}
		cell := cellForSymbol(symbol)
		if cell == domainModel.Empty {
			return fmt.Errorf("invalid symbol %q for player %s", symbol, userID)
		}
		game.Board[row][col] = cell

		game.GameState, game.WinnerID = a.gameService.CheckCompletion(game)

		if game.Mode == domainModel.ModeVSComputer && game.GameState == domainModel.StatePlaying {
			game.Board = a.gameService.ComputeNextMove(game)
			game.GameState, game.WinnerID = a.gameService.CheckCompletion(game)
		}

		if game.GameState != domainModel.StatePlaying {
			game.NextPlayerID = nil
		} else if game.Mode == domainModel.ModeVSPlayer {
			if userID == game.HostID {
				game.NextPlayerID = game.OpponentID
			} else {
				hostID := game.HostID
				game.NextPlayerID = &hostID
			}
		}

		if err := a.gameRepo.SaveGame(txCtx, game); err != nil {
			return fmt.Errorf("save game after move: %w", err)
		}
		resultGame = game
		return nil
	})

	return resultGame, err
}

func (a *appService) GetWaitingGames(ctx context.Context) ([]domainModel.Game, error) {
	games, err := a.gameRepo.GetGamesByState(ctx, domainModel.StateWaiting)
	if err != nil {
		return nil, fmt.Errorf("get waiting games: %w", err)
	}
	return games, nil
}

func (a *appService) GetFinishedGamesByUserID(ctx context.Context, userID uuid.UUID) ([]domainModel.Game, error) {
	games, err := a.gameRepo.GetCompletedGamesByPlayerID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get finished games: %w", err)
	}
	return games, nil
}

func (a *appService) GetCurrentGame(ctx context.Context, userID uuid.UUID) (domainModel.Game, error) {
	game, err := a.gameRepo.GetCurrentGame(ctx, userID)
	if err != nil {
		if errors.Is(err, storage.ErrGameNotFound) {
			return domainModel.Game{}, ErrGameNotFound
		}
		return domainModel.Game{}, fmt.Errorf("get current game: %w", err)
	}
	return game, nil
}

func (a *appService) GetLeaderboard(ctx context.Context, limit int) ([]domainModel.LeaderboardEntry, error) {
	if limit <= 0 {
		return nil, ErrInvalidLimit
	}
	entries, err := a.gameRepo.GetLeaderboard(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("get leaderboard: %w", err)
	}
	return entries, nil
}

func createEmptyBoard(size int) domainModel.Board {
	board := make(domainModel.Board, size)
	for i := range board {
		board[i] = make([]int, size)
	}
	return board
}

func cellForSymbol(symbol string) int {
	switch symbol {
	case domainModel.SymbolX:
		return domainModel.Cross
	case domainModel.SymbolO:
		return domainModel.Circle
	default:
		return domainModel.Empty
	}
}
