// internal/domain/port/game_service.go
package port

import (
	"tictactoe/internal/domain/entity"
)

type GameService interface {
	StartGame(userID string, boardSize, winningLength int) (*entity.Game, error)
	SearchPendingGames(boardSize, winningLength int) ([]*entity.Game, error)
	JoinGame(userID, gameID string) (*entity.Game, error)
	MakeMove(userID, gameID string, row, col int) (*entity.Game, error)
	GetGame(gameID, userID string) (*entity.Game, error)
	GetUserStats(userID string) (*entity.UserStats, error)
}
