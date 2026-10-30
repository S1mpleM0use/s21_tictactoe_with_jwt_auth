package service

import (
	domainModel "tictactoe/internal/domain/model"
)

func (s *gameService) minimax(game domainModel.Game, depth int, isMax bool) int {
	state, winner := s.CheckCompletion(game)
	if state != domainModel.StatePlaying {
		if winner == nil {
			return 0
		}
		switch *winner {
		case domainModel.ComputerID:
			return 10 - depth
		case game.HostID:
			return -10 + depth
		default:
			return 0
		}
	}
	if isMax {
		bestScore := -1000
		for i := 0; i < domainModel.BoardDimension; i++ {
			for j := 0; j < domainModel.BoardDimension; j++ {
				if game.Board[i][j] == domainModel.Empty {
					game.Board[i][j] = domainModel.Circle
					score := s.minimax(game, depth+1, false)
					game.Board[i][j] = domainModel.Empty

					if score > bestScore {
						bestScore = score
					}
				}
			}
		}
		return bestScore
	}

	bestScore := 1000
	for i := 0; i < domainModel.BoardDimension; i++ {
		for j := 0; j < domainModel.BoardDimension; j++ {
			if game.Board[i][j] == domainModel.Empty {
				game.Board[i][j] = domainModel.Cross
				score := s.minimax(game, depth+1, true)
				game.Board[i][j] = domainModel.Empty

				if score < bestScore {
					bestScore = score
				}
			}
		}
	}
	return bestScore
}
