package service

import (
	domainModel "tictactoe/internal/domain/model"

	"github.com/google/uuid"
)

type gameService struct{}

func NewGameService() GameService {
	return &gameService{}
}

func (s *gameService) ComputeNextMove(game domainModel.Game) domainModel.Board {
	bestScore := -1000
	bestRow := -1
	bestCol := -1

	for i := 0; i < domainModel.BoardDimension; i++ {
		for j := 0; j < domainModel.BoardDimension; j++ {
			if game.Board[i][j] == domainModel.Empty {
				game.Board[i][j] = domainModel.Circle
				score := s.minimax(game, 0, false)
				game.Board[i][j] = domainModel.Empty

				if score > bestScore {
					bestScore = score
					bestRow = i
					bestCol = j
				}
			}
		}
	}
	if bestRow != -1 && bestCol != -1 {
		game.Board[bestRow][bestCol] = domainModel.Circle
	}
	return game.Board
}

func (s *gameService) CheckCompletion(game domainModel.Game) (string, *uuid.UUID) {
	if checkForWinningRow(game.Board, domainModel.Cross) ||
		checkForWinningColumn(game.Board, domainModel.Cross) ||
		checkForWinningDiagonals(game.Board, domainModel.Cross) {
		hostID := game.HostID
		return domainModel.StateFinished, &hostID
	}

	if checkForWinningRow(game.Board, domainModel.Circle) ||
		checkForWinningColumn(game.Board, domainModel.Circle) ||
		checkForWinningDiagonals(game.Board, domainModel.Circle) {
		if game.OpponentID != nil {
			oppID := *game.OpponentID
			return domainModel.StateFinished, &oppID
		}
		return domainModel.StateFinished, nil
	}

	for i := 0; i < domainModel.BoardDimension; i++ {
		for j := 0; j < domainModel.BoardDimension; j++ {
			if game.Board[i][j] == domainModel.Empty {
				return domainModel.StatePlaying, nil
			}
		}
	}

	return domainModel.StateDraw, nil
}
