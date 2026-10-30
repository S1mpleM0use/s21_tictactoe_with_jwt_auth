package handler

import (
	"net/http"
)

func (h *Handler) Routes() *http.ServeMux {
	mux := http.NewServeMux()

	// public
	mux.HandleFunc("POST /user", h.register)
	mux.HandleFunc("POST /user/login", h.login)
	mux.HandleFunc("POST /user/refresh-access", h.refreshAccessToken)

	// authenticated only
	mux.Handle("POST /user/refresh-token", h.authenticated(h.refreshRefreshToken))
	mux.Handle("GET /user/me", h.authenticated(h.GetMe))
	mux.Handle("GET /user/{userID}", h.authenticated(h.GetUserInfoByID))

	mux.Handle("POST /game", h.authenticated(h.CreateGame))
	mux.Handle("GET /game/list", h.authenticated(h.GetWaitingGames))
	mux.Handle("GET /game/current", h.authenticated(h.GetCurrentGame))
	mux.Handle("GET /game/finished", h.authenticated(h.GetCompletedGamesByUserID))
	mux.Handle("GET /game/leaderboard", h.authenticated(h.GetLeaderboard))
	mux.Handle("GET /game/leaderboard/{limit}", h.authenticated(h.GetLeaderboard))
	mux.Handle("POST /game/{gameID}/join", h.authenticated(h.JoinGame))
	mux.Handle("POST /game/{gameID}/move", h.authenticated(h.MakeMove))

	return mux
}
