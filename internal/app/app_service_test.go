package app

import (
	"context"
	"errors"
	"testing"

	domainModel "tictactoe/internal/domain/model"
	"tictactoe/internal/domain/service"
	"tictactoe/internal/infrastructure/datasource/storage"

	"github.com/google/uuid"
)

type mockTxManager struct {
	calls      int
	committed  int
	rolledBack int
}

func (m *mockTxManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	m.calls++
	if err := fn(ctx); err != nil {
		m.rolledBack++
		return err
	}
	m.committed++
	return nil
}

type mockGameRepo struct {
	games       map[uuid.UUID]domainModel.Game
	leaderboard []domainModel.LeaderboardEntry
	saveErr     error
}

func newMockGameRepo() *mockGameRepo {
	return &mockGameRepo{
		games: make(map[uuid.UUID]domainModel.Game),
	}
}

func cloneBoard(b domainModel.Board) domainModel.Board {
	cp := make(domainModel.Board, len(b))
	for i := range b {
		cp[i] = append([]int(nil), b[i]...)
	}
	return cp
}

func cloneGame(g domainModel.Game) domainModel.Game {
	cp := g
	cp.Board = cloneBoard(g.Board)
	if g.Symbols != nil {
		cp.Symbols = make(map[string]string, len(g.Symbols))
		for k, v := range g.Symbols {
			cp.Symbols[k] = v
		}
	}
	return cp
}

func (m *mockGameRepo) SaveGame(_ context.Context, game domainModel.Game) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.games[game.ID] = cloneGame(game)
	return nil
}

func (m *mockGameRepo) GetGameByID(_ context.Context, id uuid.UUID) (domainModel.Game, error) {
	g, ok := m.games[id]
	if !ok {
		return domainModel.Game{}, storage.ErrGameNotFound
	}
	return cloneGame(g), nil
}

func (m *mockGameRepo) GetGamesByState(_ context.Context, state string) ([]domainModel.Game, error) {
	var res []domainModel.Game
	for _, g := range m.games {
		if g.GameState == state {
			res = append(res, cloneGame(g))
		}
	}
	return res, nil
}

func (m *mockGameRepo) GetCompletedGamesByPlayerID(_ context.Context, playerID uuid.UUID) ([]domainModel.Game, error) {
	var res []domainModel.Game
	for _, g := range m.games {
		isParticipant := g.HostID == playerID || (g.OpponentID != nil && *g.OpponentID == playerID)
		isFinished := g.GameState == domainModel.StateFinished || g.GameState == domainModel.StateDraw
		if isParticipant && isFinished {
			res = append(res, cloneGame(g))
		}
	}
	return res, nil
}

func (m *mockGameRepo) GetCurrentGame(_ context.Context, userID uuid.UUID) (domainModel.Game, error) {
	for _, g := range m.games {
		isParticipant := g.HostID == userID || (g.OpponentID != nil && *g.OpponentID == userID)
		isActive := g.GameState == domainModel.StateWaiting || g.GameState == domainModel.StatePlaying
		if isParticipant && isActive {
			return cloneGame(g), nil
		}
	}
	return domainModel.Game{}, storage.ErrGameNotFound
}

func (m *mockGameRepo) GetLeaderboard(_ context.Context, limit int) ([]domainModel.LeaderboardEntry, error) {
	if limit > len(m.leaderboard) {
		limit = len(m.leaderboard)
	}
	return m.leaderboard[:limit], nil
}

func TestAppService_CreateGame(t *testing.T) {
	ctx := context.Background()
	repo := newMockGameRepo()
	tx := &mockTxManager{}
	svc := NewAppService(service.NewGameService(), repo, tx)
	hostID := uuid.New()

	t.Run("invalid mode returns ErrInvalidMode", func(t *testing.T) {
		_, err := svc.CreateGame(ctx, hostID, "invalid_mode")
		if !errors.Is(err, ErrInvalidMode) {
			t.Fatalf("expected ErrInvalidMode, got %v", err)
		}
	})

	t.Run("VScomputer initializes playing state, opponent, symbols and CreatedAt", func(t *testing.T) {
		game, err := svc.CreateGame(ctx, hostID, domainModel.ModeVSComputer)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if game.GameState != domainModel.StatePlaying {
			t.Fatalf("expected state %s, got %s", domainModel.StatePlaying, game.GameState)
		}
		if game.OpponentID == nil || *game.OpponentID != domainModel.ComputerID {
			t.Fatalf("expected computer opponent, got %v", game.OpponentID)
		}
		if game.CreatedAt.IsZero() {
			t.Fatal("expected non-zero CreatedAt")
		}
		if game.Symbols[hostID.String()] != domainModel.SymbolX || game.Symbols[domainModel.ComputerID.String()] != domainModel.SymbolO {
			t.Fatalf("unexpected symbols map: %+v", game.Symbols)
		}
	})

	t.Run("VSplayer initializes waiting state without opponent", func(t *testing.T) {
		game, err := svc.CreateGame(ctx, hostID, domainModel.ModeVSPlayer)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if game.GameState != domainModel.StateWaiting || game.OpponentID != nil {
			t.Fatalf("unexpected VSplayer initial state: %+v", game)
		}
	})
}

func TestAppService_JoinGame_WithTransactions(t *testing.T) {
	ctx := context.Background()
	repo := newMockGameRepo()
	tx := &mockTxManager{}
	svc := NewAppService(service.NewGameService(), repo, tx)

	hostID := uuid.New()
	guestID := uuid.New()

	waitingGame, err := svc.CreateGame(ctx, hostID, domainModel.ModeVSPlayer)
	if err != nil {
		t.Fatalf("create waiting game: %v", err)
	}

	t.Run("creator cannot join own game and transaction rolls back", func(t *testing.T) {
		_, err := svc.JoinGame(ctx, waitingGame.ID, hostID)
		if !errors.Is(err, ErrCreatorCannotJoin) {
			t.Fatalf("expected ErrCreatorCannotJoin, got %v", err)
		}
		if tx.rolledBack != 1 {
			t.Fatalf("expected 1 rollback, got %d", tx.rolledBack)
		}
	})

	t.Run("second player successfully joins inside transaction", func(t *testing.T) {
		joined, err := svc.JoinGame(ctx, waitingGame.ID, guestID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if joined.GameState != domainModel.StatePlaying {
			t.Fatalf("expected state playing, got %s", joined.GameState)
		}
		if joined.OpponentID == nil || *joined.OpponentID != guestID {
			t.Fatalf("expected opponent %s, got %v", guestID, joined.OpponentID)
		}
		if joined.NextPlayerID == nil || *joined.NextPlayerID != hostID {
			t.Fatalf("expected next player %s, got %v", hostID, joined.NextPlayerID)
		}
		if tx.committed != 1 {
			t.Fatalf("expected 1 commit, got %d", tx.committed)
		}
	})

	t.Run("joining already started game returns ErrNotWaiting", func(t *testing.T) {
		_, err := svc.JoinGame(ctx, waitingGame.ID, uuid.New())
		if !errors.Is(err, ErrNotWaiting) {
			t.Fatalf("expected ErrNotWaiting, got %v", err)
		}
	})
}

func TestAppService_ApplyMove_MultiplayerAndComputer(t *testing.T) {
	ctx := context.Background()
	repo := newMockGameRepo()
	tx := &mockTxManager{}
	svc := NewAppService(service.NewGameService(), repo, tx)

	hostID := uuid.New()
	guestID := uuid.New()

	t.Run("VScomputer move triggers AI response and preserves CreatedAt", func(t *testing.T) {
		game, _ := svc.CreateGame(ctx, hostID, domainModel.ModeVSComputer)
		updated, err := svc.ApplyMove(ctx, game.ID, hostID, 0, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Board[0][0] != domainModel.Cross {
			t.Fatalf("expected Cross at (0,0), got %d", updated.Board[0][0])
		}
		// Проверяем, что бот тоже сделал ход (Circle)
		circles := 0
		for r := 0; r < domainModel.BoardDimension; r++ {
			for c := 0; c < domainModel.BoardDimension; c++ {
				if updated.Board[r][c] == domainModel.Circle {
					circles++
				}
			}
		}
		if circles != 1 {
			t.Fatalf("expected computer to place 1 Circle, found %d", circles)
		}
		if updated.CreatedAt.IsZero() {
			t.Fatal("expected CreatedAt to be preserved after move")
		}
	})

	t.Run("VSplayer full winning flow and move validations", func(t *testing.T) {
		game, _ := svc.CreateGame(ctx, hostID, domainModel.ModeVSPlayer)
		_, _ = svc.JoinGame(ctx, game.ID, guestID)

		// Чужой игрок
		if _, err := svc.ApplyMove(ctx, game.ID, uuid.New(), 0, 0); !errors.Is(err, ErrNotYourGame) {
			t.Fatalf("expected ErrNotYourGame, got %v", err)
		}

		// Не своя очередь (сейчас ход hostID)
		if _, err := svc.ApplyMove(ctx, game.ID, guestID, 0, 0); !errors.Is(err, ErrNotYourTurn) {
			t.Fatalf("expected ErrNotYourTurn, got %v", err)
		}

		// Выход за границы поля
		if _, err := svc.ApplyMove(ctx, game.ID, hostID, -1, 0); !errors.Is(err, ErrCellOutOfBounds) {
			t.Fatalf("expected ErrCellOutOfBounds, got %v", err)
		}
		if _, err := svc.ApplyMove(ctx, game.ID, hostID, 0, 3); !errors.Is(err, ErrCellOutOfBounds) {
			t.Fatalf("expected ErrCellOutOfBounds, got %v", err)
		}

		// Ход 1: Host (0,0)
		if _, err := svc.ApplyMove(ctx, game.ID, hostID, 0, 0); err != nil {
			t.Fatalf("move 1 failed: %v", err)
		}

		// Попытка сходить в занятую клетку
		if _, err := svc.ApplyMove(ctx, game.ID, guestID, 0, 0); !errors.Is(err, ErrCellOccupied) {
			t.Fatalf("expected ErrCellOccupied, got %v", err)
		}

		// Разыгрываем победу для Host по первой строке: (0,0), (0,1), (0,2)
		_, _ = svc.ApplyMove(ctx, game.ID, guestID, 1, 0)
		_, _ = svc.ApplyMove(ctx, game.ID, hostID, 0, 1)
		_, _ = svc.ApplyMove(ctx, game.ID, guestID, 1, 1)
		finalGame, err := svc.ApplyMove(ctx, game.ID, hostID, 0, 2)
		if err != nil {
			t.Fatalf("winning move failed: %v", err)
		}

		if finalGame.GameState != domainModel.StateFinished {
			t.Fatalf("expected state finished, got %s", finalGame.GameState)
		}
		if finalGame.WinnerID == nil || *finalGame.WinnerID != hostID {
			t.Fatalf("expected winner %s, got %v", hostID, finalGame.WinnerID)
		}
		if finalGame.NextPlayerID != nil {
			t.Fatalf("expected nil NextPlayerID after finish, got %v", finalGame.NextPlayerID)
		}

		// Ход после завершения игры запрещен
		if _, err := svc.ApplyMove(ctx, game.ID, guestID, 2, 2); !errors.Is(err, ErrGameNotPlaying) {
			t.Fatalf("expected ErrGameNotPlaying, got %v", err)
		}
	})
}

func TestAppService_LeaderboardAndQueries(t *testing.T) {
	ctx := context.Background()
	repo := newMockGameRepo()
	repo.leaderboard = []domainModel.LeaderboardEntry{
		{UserID: uuid.New(), Login: "pro", Wins: 10, Losses: 2, Draws: 1, WinRatio: 0.769},
		{UserID: uuid.New(), Login: "rookie", Wins: 1, Losses: 3, Draws: 0, WinRatio: 0.25},
	}
	svc := NewAppService(service.NewGameService(), repo, &mockTxManager{})

	if _, err := svc.GetLeaderboard(ctx, 0); !errors.Is(err, ErrInvalidLimit) {
		t.Fatalf("expected ErrInvalidLimit for limit=0, got %v", err)
	}
	if _, err := svc.GetLeaderboard(ctx, -5); !errors.Is(err, ErrInvalidLimit) {
		t.Fatalf("expected ErrInvalidLimit for limit=-5, got %v", err)
	}

	top1, err := svc.GetLeaderboard(ctx, 1)
	if err != nil || len(top1) != 1 || top1[0].Login != "pro" {
		t.Fatalf("unexpected leaderboard result: %+v, err=%v", top1, err)
	}
}
