package mapper

import (
	domainModel "tictactoe/internal/domain/model"
	webModel "tictactoe/internal/web/model"
)

func ToGameResponse(g domainModel.Game) webModel.GameResponse {
	return webModel.GameResponse{
		ID:           g.ID,
		Board:        g.Board,
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

func UserToWeb(u domainModel.User) webModel.User {
	return webModel.User{ID: u.ID, Login: u.Login}
}

func JwtToWeb(j domainModel.JwtResponse) webModel.JwtResponse {
	return webModel.JwtResponse{
		Type:         j.Type,
		AccessToken:  j.AccessToken,
		RefreshToken: j.RefreshToken,
	}
}

func ToLeaderboardResponse(e domainModel.LeaderboardEntry) webModel.LeaderboardEntryResponse {
	return webModel.LeaderboardEntryResponse{
		UserID:   e.UserID,
		Login:    e.Login,
		Wins:     e.Wins,
		Losses:   e.Losses,
		Draws:    e.Draws,
		WinRatio: e.WinRatio,
	}
}
