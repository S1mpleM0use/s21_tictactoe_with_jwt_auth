package model

import (
	"time"

	"github.com/google/uuid"
)

type GameResponse struct {
	ID           uuid.UUID         `json:"id"`
	Board        [][]int           `json:"board"`
	GameState    string            `json:"game_state"` // "waiting" | "playing" | "finished" | "draw"
	Mode         string            `json:"mode"`       // "VSplayer" | "VScomputer"
	HostID       uuid.UUID         `json:"host_id"`
	OpponentID   *uuid.UUID        `json:"opponent_id,omitempty"`
	NextPlayerID *uuid.UUID        `json:"next_player_id,omitempty"`
	Symbols      map[string]string `json:"symbols,omitempty"`
	WinnerID     *uuid.UUID        `json:"winner_id,omitempty"`
	CreatedAt    time.Time         `json:"created_at"`
}

type MoveRequest struct {
	Row    *int `json:"row"`
	Column *int `json:"column"`
}

type CreateGameRequest struct {
	Mode string `json:"mode"` // "VSplayer" или "VScomputer"
}

type SignUpRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type JwtRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type JwtResponse struct {
	Type         string `json:"type"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type RefreshJwtRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type User struct {
	ID    uuid.UUID `json:"id"`
	Login string    `json:"login"`
}

type LeaderboardEntryResponse struct {
	UserID   uuid.UUID `json:"user_id"`
	Login    string    `json:"login"`
	Wins     int       `json:"wins"`
	Losses   int       `json:"losses"`
	Draws    int       `json:"draws"`
	WinRatio float64   `json:"win_ratio"`
}
