package main

import (
	"tictactoe/internal/di"

	"go.uber.org/fx"
)

func main() {
	fx.New(di.Module).Run()
}
