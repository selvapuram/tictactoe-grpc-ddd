// internal/adapters/repository/in_memory_user_repository.go
package repository

import (
	"sync"
	"tictactoe/internal/domain/entity"
	"tictactoe/internal/domain/port"
)

type inMemoryUserRepository struct {
	mu    sync.RWMutex
	users map[string]*entity.UserStats
}

func NewInMemoryUserRepository() port.UserRepository {
	return &inMemoryUserRepository{
		users: make(map[string]*entity.UserStats),
	}
}

func (r *inMemoryUserRepository) SaveStats(stats *entity.UserStats) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Deep copy
	statsCopy := *stats
	r.users[stats.UserID] = &statsCopy
	return nil
}

func (r *inMemoryUserRepository) FindStatsByUserID(userID string) (*entity.UserStats, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	stats, exists := r.users[userID]
	if !exists {
		return nil, entity.ErrUserNotFound
	}
	
	// Deep copy
	statsCopy := *stats
	return &statsCopy, nil
}

func (r *inMemoryUserRepository) CreateUserIfNotExists(userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.users[userID]; !exists {
		r.users[userID] = entity.NewUserStats(userID)
	}
	
	return nil
}
