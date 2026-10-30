package model

import "github.com/google/uuid"

const (
	BoardDimension = 3
	Empty          = 0
	Cross          = 1
	Circle         = 2
)

const (
	SymbolX = "X"
	SymbolO = "O"
)

const (
	ModeVSComputer = "VScomputer"
	ModeVSPlayer   = "VSplayer"
)

const (
	StateWaiting  = "waiting"
	StatePlaying  = "playing"
	StateDraw     = "draw"
	StateFinished = "finished"
)

var (
	ComputerID = uuid.MustParse("00000000-0000-0000-0000-000000000001")
)
