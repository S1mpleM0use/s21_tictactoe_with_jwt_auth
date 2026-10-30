package model

import (
	"time"

	"github.com/google/uuid"
)

type GameDTO struct {
	ID           uuid.UUID         `db:"id"`
	Board        BoardDTO          `db:"board"`
	GameState    string            `db:"state"`
	Mode         string            `db:"mode"`
	HostID       uuid.UUID         `db:"host_id"`
	OpponentID   *uuid.UUID        `db:"opp_id"`
	NextPlayerID *uuid.UUID        `db:"next_id"`
	Symbols      map[string]string `db:"symbols"`
	WinnerID     *uuid.UUID        `db:"winner_id"`
	CreatedAt    time.Time         `db:"created_at"`
}

type LeaderboardEntryDTO struct {
	UserID   uuid.UUID `db:"user_id"`
	Login    string    `db:"login"`
	Wins     int       `db:"wins"`
	Losses   int       `db:"losses"`
	Draws    int       `db:"draws"`
	WinRatio float64   `db:"win_ratio"`
}
