package pgsql

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	infraModel "tictactoe/internal/infrastructure/datasource/model"
)

func TestGameStorage_IntegrationAndErrors(t *testing.T) {
	cfg := NewConfig()
	storage, err := NewGameStorage(cfg)
	if err != nil {
		t.Fatalf("Failed to connect to DB: %v. Make sure start_psql.sh is running", err)
	}
	defer storage.Close()

	ctx := context.Background()

	if err := Migrate(ctx, storage); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	userID := uuid.New()
	uniqueLogin := "test_" + userID.String()[:12]
	testUser := infraModel.UserDTO{
		ID:           userID,
		Login:        uniqueLogin,
		PasswordHash: "hashed_pass",
	}

	t.Run("SaveUser and duplicate login constraint (ErrLoginAlreadyTaken)", func(t *testing.T) {
		if err := storage.SaveUser(ctx, testUser); err != nil {
			t.Fatalf("Failed to save user: %v", err)
		}

		duplicateUser := infraModel.UserDTO{
			ID:           uuid.New(),
			Login:        uniqueLogin,
			PasswordHash: "other_hash",
		}
		err := storage.SaveUser(ctx, duplicateUser)
		if !errors.Is(err, ErrLoginAlreadyTaken) {
			t.Fatalf("Expected ErrLoginAlreadyTaken on duplicate login, got: %v", err)
		}
	})

	t.Run("GetUserByLogin and GetUserByID success and ErrUserNotFound", func(t *testing.T) {
		byLogin, err := storage.GetUserByLogin(ctx, uniqueLogin)
		if err != nil || byLogin.ID != userID {
			t.Fatalf("GetUserByLogin failed: user=%+v, err=%v", byLogin, err)
		}

		byID, err := storage.GetUserByID(ctx, userID)
		if err != nil || byID.Login != uniqueLogin {
			t.Fatalf("GetUserByID failed: user=%+v, err=%v", byID, err)
		}

		_, err = storage.GetUserByLogin(ctx, "non_existent_"+uuid.NewString()[:8])
		if !errors.Is(err, ErrUserNotFound) {
			t.Fatalf("Expected ErrUserNotFound by login, got: %v", err)
		}

		_, err = storage.GetUserByID(ctx, uuid.New())
		if !errors.Is(err, ErrUserNotFound) {
			t.Fatalf("Expected ErrUserNotFound by ID, got: %v", err)
		}
	})

	t.Run("StoreGame, LoadGameByID, LoadCurrentGame, CreatedAt and Leaderboard", func(t *testing.T) {
		_, err := storage.LoadGameByID(ctx, uuid.New())
		if !errors.Is(err, ErrGameNotFound) {
			t.Fatalf("Expected ErrGameNotFound for unknown ID, got: %v", err)
		}

		_, err = storage.LoadCurrentGame(ctx, uuid.New())
		if !errors.Is(err, ErrGameNotFound) {
			t.Fatalf("Expected ErrGameNotFound for user without games, got: %v", err)
		}

		gameID := uuid.New()
		now := time.Now().UTC().Truncate(time.Microsecond)
		game := infraModel.GameDTO{
			ID:           gameID,
			Board:        "000010000",
			GameState:    "waiting",
			Mode:         "VSplayer",
			HostID:       userID,
			OpponentID:   nil,
			NextPlayerID: &userID,
			Symbols:      map[string]string{userID.String(): "X"},
			WinnerID:     nil,
			CreatedAt:    now,
		}

		if err := storage.StoreGame(ctx, game); err != nil {
			t.Fatalf("Failed to store game: %v", err)
		}

		loadedGame, err := storage.LoadGameByID(ctx, gameID)
		if err != nil {
			t.Fatalf("Failed to load game: %v", err)
		}
		if loadedGame.Board != game.Board || loadedGame.GameState != game.GameState || loadedGame.HostID != game.HostID {
			t.Fatalf("Loaded game mismatch: got %+v, want %+v", loadedGame, game)
		}
		if loadedGame.CreatedAt.IsZero() {
			t.Fatal("Expected non-zero CreatedAt from LoadGameByID")
		}

		currentGame, err := storage.LoadCurrentGame(ctx, userID)
		if err != nil || currentGame.ID != gameID || currentGame.CreatedAt.IsZero() {
			t.Fatalf("LoadCurrentGame failed: game=%+v, err=%v", currentGame, err)
		}

		// Завершаем игру победой testUser и проверяем лидерборд
		game.GameState = "finished"
		game.WinnerID = &userID
		if err := storage.StoreGame(ctx, game); err != nil {
			t.Fatalf("Failed to update game to finished: %v", err)
		}

		leaderboard, err := storage.LoadLeaderboard(ctx, 10)
		if err != nil {
			t.Fatalf("LoadLeaderboard failed: %v", err)
		}
		if len(leaderboard) == 0 {
			t.Fatal("Expected non-empty leaderboard")
		}
	})

	t.Run("WithinTransaction commits on success and rolls back on error", func(t *testing.T) {
		txGameID := uuid.New()
		txGame := infraModel.GameDTO{
			ID:        txGameID,
			Board:     "000000000",
			GameState: "waiting",
			Mode:      "VSplayer",
			HostID:    userID,
			Symbols:   map[string]string{userID.String(): "X"},
			CreatedAt: time.Now().UTC(),
		}

		// 1. Проверяем Rollback при возврате ошибки
		simulatedErr := errors.New("simulated failure")
		err := storage.WithinTransaction(ctx, func(txCtx context.Context) error {
			if err := storage.StoreGame(txCtx, txGame); err != nil {
				return err
			}
			return simulatedErr
		})
		if !errors.Is(err, simulatedErr) {
			t.Fatalf("Expected simulatedErr, got: %v", err)
		}

		// Игры не должно быть в БД после отката
		if _, err := storage.LoadGameByID(ctx, txGameID); !errors.Is(err, ErrGameNotFound) {
			t.Fatalf("Expected ErrGameNotFound after rollback, got: %v", err)
		}

		// 2. Проверяем Commit при успешном выполнении
		err = storage.WithinTransaction(ctx, func(txCtx context.Context) error {
			if err := storage.StoreGame(txCtx, txGame); err != nil {
				return err
			}
			loaded, err := storage.LoadGameByID(txCtx, txGameID) // здесь сработает SELECT ... FOR UPDATE
			if err != nil || loaded.ID != txGameID {
				return errors.New("failed to load game inside tx")
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Expected successful commit, got: %v", err)
		}

		// Теперь игра должна быть в БД
		if _, err := storage.LoadGameByID(ctx, txGameID); err != nil {
			t.Fatalf("Expected game to persist after commit, got: %v", err)
		}
	})
}
