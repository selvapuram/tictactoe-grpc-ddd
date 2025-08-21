// internal/domain/port/game_repository.go
package port

import "tictactoe/internal/domain/entity"

type GameRepository interface {
	Save(game *entity.Game) error
	FindByID(id string) (*entity.Game, error)
	FindPendingGames(boardSize, winningLength int) ([]*entity.Game, error)
	Delete(id string) error
	Count() int64
}
