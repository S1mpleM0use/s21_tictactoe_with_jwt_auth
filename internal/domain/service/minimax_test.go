package service

import (
	"testing"
	domainModel "tictactoe/internal/domain/model"

	"github.com/google/uuid"
)

func TestMinimaxScores(t *testing.T) {
	svc := &gameService{}
	hostID := uuid.New()
	compID := domainModel.ComputerID

	tests := []struct {
		name      string
		board     domainModel.Board
		isMax     bool
		wantScore int
	}{
		{
			name: "Computer (Circle) is about to win",
			board: domainModel.Board{
				{domainModel.Circle, domainModel.Circle, domainModel.Empty},
				{domainModel.Cross, domainModel.Cross, domainModel.Empty},
				{domainModel.Empty, domainModel.Empty, domainModel.Empty},
			},
			isMax:     true,
			wantScore: 9,
		},
		{
			name: "Player (Cross) is about to win, Computer must block",
			board: domainModel.Board{
				{domainModel.Cross, domainModel.Cross, domainModel.Empty},
				{domainModel.Circle, domainModel.Empty, domainModel.Empty},
				{domainModel.Empty, domainModel.Empty, domainModel.Empty},
			},
			isMax:     false,
			wantScore: -9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := domainModel.Game{
				Board:      tt.board,
				HostID:     hostID,
				OpponentID: &compID,
			}

			score := svc.minimax(game, 0, tt.isMax)

			if (tt.wantScore > 0 && score <= 0) || (tt.wantScore < 0 && score >= 0) {
				t.Errorf("Expected score around %d, got %d", tt.wantScore, score)
			}
		})
	}
}

func TestComputeNextMove_BlocksPlayerWinningMove(t *testing.T) {
	svc := NewGameService()
	compID := domainModel.ComputerID
	game := domainModel.Game{
		HostID:     uuid.New(),
		OpponentID: &compID,
		Board: domainModel.Board{
			{domainModel.Cross, domainModel.Cross, domainModel.Empty},
			{domainModel.Empty, domainModel.Circle, domainModel.Empty},
			{domainModel.Empty, domainModel.Empty, domainModel.Empty},
		},
	}

	nextBoard := svc.ComputeNextMove(game)
	if nextBoard[0][2] != domainModel.Circle {
		t.Fatalf("expected computer to block player at (0,2), got cell=%d", nextBoard[0][2])
	}
}

func TestComputeNextMove_FullBoardUnchanged(t *testing.T) {
	svc := NewGameService()
	compID := domainModel.ComputerID
	game := domainModel.Game{
		HostID:     uuid.New(),
		OpponentID: &compID,
		Board: domainModel.Board{
			{domainModel.Cross, domainModel.Circle, domainModel.Cross},
			{domainModel.Cross, domainModel.Circle, domainModel.Circle},
			{domainModel.Circle, domainModel.Cross, domainModel.Cross},
		},
	}

	nextBoard := svc.ComputeNextMove(game)
	for r := 0; r < domainModel.BoardDimension; r++ {
		for c := 0; c < domainModel.BoardDimension; c++ {
			if nextBoard[r][c] != game.Board[r][c] {
				t.Fatalf("expected full board to remain unchanged at (%d,%d)", r, c)
			}
		}
	}
}

func TestComputeNextMove_ComputerNeverLosesFromAnyOpening(t *testing.T) {
	svc := NewGameService()
	hostID := uuid.New()
	compID := domainModel.ComputerID

	// Проверяем все 9 возможных первых ходов игрока (Cross)
	for startRow := 0; startRow < domainModel.BoardDimension; startRow++ {
		for startCol := 0; startCol < domainModel.BoardDimension; startCol++ {
			board := domainModel.Board{
				{domainModel.Empty, domainModel.Empty, domainModel.Empty},
				{domainModel.Empty, domainModel.Empty, domainModel.Empty},
				{domainModel.Empty, domainModel.Empty, domainModel.Empty},
			}
			board[startRow][startCol] = domainModel.Cross

			game := domainModel.Game{
				HostID:     hostID,
				OpponentID: &compID,
				Board:      board,
			}

			// Симулируем партию: бот отвечает оптимально, игрок ходит в первую свободную клетку
			for {
				state, winner := svc.CheckCompletion(game)
				if state != domainModel.StatePlaying {
					if winner != nil && *winner == hostID {
						t.Fatalf("computer lost starting from opening (%d,%d)", startRow, startCol)
					}
					break
				}

				// Ход компьютера
				game.Board = svc.ComputeNextMove(game)
				state, winner = svc.CheckCompletion(game)
				if state != domainModel.StatePlaying {
					if winner != nil && *winner == hostID {
						t.Fatalf("computer lost starting from opening (%d,%d)", startRow, startCol)
					}
					break
				}

				// Ход игрока в первую свободную ячейку
				moved := false
				for r := 0; r < domainModel.BoardDimension && !moved; r++ {
					for c := 0; c < domainModel.BoardDimension && !moved; c++ {
						if game.Board[r][c] == domainModel.Empty {
							game.Board[r][c] = domainModel.Cross
							moved = true
						}
					}
				}
			}
		}
	}
}
