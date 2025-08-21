// internal/domain/port/user_repository.go
package port

import "tictactoe/internal/domain/entity"

type UserRepository interface {
	SaveStats(stats *entity.UserStats) error
	FindStatsByUserID(userID string) (*entity.UserStats, error)
	CreateUserIfNotExists(userID string) error
}
