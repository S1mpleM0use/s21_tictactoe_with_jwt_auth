package mapper

import (
	"strings"
	domainModel "tictactoe/internal/domain/model"
	infraModel "tictactoe/internal/infrastructure/datasource/model"
)

func ToDomain(m infraModel.GameDTO) domainModel.Game {
	return domainModel.Game{
		ID:           m.ID,
		Board:        toDomainParser(m.Board),
		GameState:    m.GameState,
		Mode:         m.Mode,
		HostID:       m.HostID,
		OpponentID:   m.OpponentID,
		NextPlayerID: m.NextPlayerID,
		Symbols:      m.Symbols,
		WinnerID:     m.WinnerID,
		CreatedAt:    m.CreatedAt,
	}
}

func ToModel(g domainModel.Game) infraModel.GameDTO {
	return infraModel.GameDTO{
		ID:           g.ID,
		Board:        toModelParser(g.Board),
		GameState:    g.GameState,
		Mode:         g.Mode,
		HostID:       g.HostID,
		OpponentID:   g.OpponentID,
		NextPlayerID: g.NextPlayerID,
		Symbols:      g.Symbols,
		WinnerID:     g.WinnerID,
		CreatedAt:    g.CreatedAt,
	}
}

func toModelParser(board domainModel.Board) infraModel.BoardDTO {
	var sb strings.Builder
	sb.Grow(domainModel.BoardDimension * domainModel.BoardDimension)

	for i := 0; i < domainModel.BoardDimension; i++ {
		for j := 0; j < domainModel.BoardDimension; j++ {
			sb.WriteByte(byte(board[i][j]) + '0')
		}
	}

	return infraModel.BoardDTO(sb.String())
}

func toDomainParser(board infraModel.BoardDTO) domainModel.Board {
	newModelBoard := make(domainModel.Board, domainModel.BoardDimension)
	for i := range newModelBoard {
		newModelBoard[i] = make([]int, domainModel.BoardDimension)
	}

	for i := 0; i < len(board) && i < domainModel.BoardDimension*domainModel.BoardDimension; i++ {
		row := i / domainModel.BoardDimension
		col := i % domainModel.BoardDimension
		newModelBoard[row][col] = int(board[i] - '0')
	}

	return newModelBoard
}

func InfraUserToDomainUser(iUser infraModel.UserDTO) domainModel.User {
	return domainModel.User{
		ID:       iUser.ID,
		Login:    iUser.Login,
		Password: iUser.PasswordHash,
	}
}

func DomainUserToInfraUser(dUser domainModel.User) infraModel.UserDTO {
	return infraModel.UserDTO{
		ID:           dUser.ID,
		Login:        dUser.Login,
		PasswordHash: dUser.Password,
	}
}

func InfraLeaderboardToDomain(dto infraModel.LeaderboardEntryDTO) domainModel.LeaderboardEntry {
	return domainModel.LeaderboardEntry{
		UserID:   dto.UserID,
		Login:    dto.Login,
		Wins:     dto.Wins,
		Losses:   dto.Losses,
		Draws:    dto.Draws,
		WinRatio: dto.WinRatio,
	}
}
