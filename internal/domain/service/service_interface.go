package service

import (
	domainModel "tictactoe/internal/domain/model"

	"github.com/google/uuid"
)

type GameService interface {
	ComputeNextMove(game domainModel.Game) domainModel.Board
	CheckCompletion(game domainModel.Game) (string, *uuid.UUID)
}
