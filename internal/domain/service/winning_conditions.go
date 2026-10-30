package service

import domainModel "tictactoe/internal/domain/model"

func checkForWinningRow(board domainModel.Board, element int) bool {
	for i := 0; i < domainModel.BoardDimension; i++ {
		rowWins := true
		for j := 0; j < domainModel.BoardDimension; j++ {
			if board[i][j] != element {
				rowWins = false
				break
			}
		}
		if rowWins {
			return true
		}
	}
	return false
}

func checkForWinningColumn(board domainModel.Board, element int) bool {
	for i := 0; i < domainModel.BoardDimension; i++ {
		columnWins := true
		for j := 0; j < domainModel.BoardDimension; j++ {
			if board[j][i] != element {
				columnWins = false
				break
			}
		}
		if columnWins {
			return true
		}
	}
	return false
}

func checkForWinningDiagonals(board domainModel.Board, element int) bool {
	mainDiagWins := true
	sideDiagWins := true

	for i := 0; i < domainModel.BoardDimension; i++ {
		// Главная диагональ: [0][0], [1][1], [2][2]
		if board[i][i] != element {
			mainDiagWins = false
		}
		// Побочная диагональ: [0][2], [1][1], [2][0]
		if board[i][domainModel.BoardDimension-1-i] != element {
			sideDiagWins = false
		}
	}

	return mainDiagWins || sideDiagWins
}
