package service

import (
	"testing"
	domainModel "tictactoe/internal/domain/model"

	"github.com/google/uuid"
)

func TestCheckForWinningRow(t *testing.T) {
	board := domainModel.Board{
		{domainModel.Cross, domainModel.Cross, domainModel.Cross},
		{domainModel.Empty, domainModel.Empty, domainModel.Empty},
		{domainModel.Empty, domainModel.Empty, domainModel.Empty},
	}
	if !checkForWinningRow(board, domainModel.Cross) {
		t.Error("Expected winning row for Cross, got false")
	}

	board = domainModel.Board{
		{domainModel.Empty, domainModel.Empty, domainModel.Empty},
		{domainModel.Cross, domainModel.Empty, domainModel.Cross},
		{domainModel.Empty, domainModel.Empty, domainModel.Empty},
	}
	if checkForWinningRow(board, domainModel.Cross) {
		t.Error("Expected false, but got winning row")
	}
}

func TestCheckForWinningColumn(t *testing.T) {
	board := domainModel.Board{
		{domainModel.Circle, domainModel.Empty, domainModel.Empty},
		{domainModel.Circle, domainModel.Empty, domainModel.Empty},
		{domainModel.Circle, domainModel.Empty, domainModel.Empty},
	}
	if !checkForWinningColumn(board, domainModel.Circle) {
		t.Error("Expected winning column for Circle, got false")
	}
}

func TestCheckForWinningDiagonals(t *testing.T) {
	boardMain := domainModel.Board{
		{domainModel.Cross, domainModel.Empty, domainModel.Empty},
		{domainModel.Empty, domainModel.Cross, domainModel.Empty},
		{domainModel.Empty, domainModel.Empty, domainModel.Cross},
	}
	if !checkForWinningDiagonals(boardMain, domainModel.Cross) {
		t.Error("Expected main diagonal win for Cross")
	}

	boardSide := domainModel.Board{
		{domainModel.Empty, domainModel.Empty, domainModel.Circle},
		{domainModel.Empty, domainModel.Circle, domainModel.Empty},
		{domainModel.Circle, domainModel.Empty, domainModel.Empty},
	}
	if !checkForWinningDiagonals(boardSide, domainModel.Circle) {
		t.Error("Expected side diagonal win for Circle")
	}
}

func TestCheckCompletion_States(t *testing.T) {
	svc := NewGameService()
	hostID := uuid.New()
	oppID := uuid.New()

	t.Run("StatePlaying when board has empty cells and no winner", func(t *testing.T) {
		game := domainModel.Game{
			HostID:     hostID,
			OpponentID: &oppID,
			Board: domainModel.Board{
				{domainModel.Cross, domainModel.Circle, domainModel.Empty},
				{domainModel.Empty, domainModel.Cross, domainModel.Empty},
				{domainModel.Empty, domainModel.Empty, domainModel.Empty},
			},
		}
		state, winner := svc.CheckCompletion(game)
		if state != domainModel.StatePlaying || winner != nil {
			t.Fatalf("expected (%s, nil), got (%s, %v)", domainModel.StatePlaying, state, winner)
		}
	})

	t.Run("StateFinished and HostID when Cross wins", func(t *testing.T) {
		game := domainModel.Game{
			HostID:     hostID,
			OpponentID: &oppID,
			Board: domainModel.Board{
				{domainModel.Cross, domainModel.Cross, domainModel.Cross},
				{domainModel.Circle, domainModel.Circle, domainModel.Empty},
				{domainModel.Empty, domainModel.Empty, domainModel.Empty},
			},
		}
		state, winner := svc.CheckCompletion(game)
		if state != domainModel.StateFinished || winner == nil || *winner != hostID {
			t.Fatalf("expected (%s, %s), got (%s, %v)", domainModel.StateFinished, hostID, state, winner)
		}
	})

	t.Run("StateFinished and OpponentID when Circle wins", func(t *testing.T) {
		game := domainModel.Game{
			HostID:     hostID,
			OpponentID: &oppID,
			Board: domainModel.Board{
				{domainModel.Cross, domainModel.Cross, domainModel.Empty},
				{domainModel.Circle, domainModel.Circle, domainModel.Circle},
				{domainModel.Cross, domainModel.Empty, domainModel.Empty},
			},
		}
		state, winner := svc.CheckCompletion(game)
		if state != domainModel.StateFinished || winner == nil || *winner != oppID {
			t.Fatalf("expected (%s, %s), got (%s, %v)", domainModel.StateFinished, oppID, state, winner)
		}
	})

	t.Run("StateDraw and nil winner when board is full without winner", func(t *testing.T) {
		game := domainModel.Game{
			HostID:     hostID,
			OpponentID: &oppID,
			Board: domainModel.Board{
				{domainModel.Cross, domainModel.Circle, domainModel.Cross},
				{domainModel.Cross, domainModel.Circle, domainModel.Circle},
				{domainModel.Circle, domainModel.Cross, domainModel.Cross},
			},
		}
		state, winner := svc.CheckCompletion(game)
		if state != domainModel.StateDraw || winner != nil {
			t.Fatalf("expected (%s, nil), got (%s, %v)", domainModel.StateDraw, state, winner)
		}
	})
}

func TestComputeNextMove_TakesWinningCell(t *testing.T) {
	svc := NewGameService()
	compID := domainModel.ComputerID
	game := domainModel.Game{
		HostID:     uuid.New(),
		OpponentID: &compID,
		Board: domainModel.Board{
			{domainModel.Circle, domainModel.Circle, domainModel.Empty},
			{domainModel.Cross, domainModel.Cross, domainModel.Empty},
			{domainModel.Empty, domainModel.Empty, domainModel.Empty},
		},
	}

	nextBoard := svc.ComputeNextMove(game)
	if nextBoard[0][2] != domainModel.Circle {
		t.Fatalf("expected computer to play winning move at (0,2), got %d", nextBoard[0][2])
	}
}
