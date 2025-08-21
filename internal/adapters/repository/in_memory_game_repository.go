// internal/adapters/repository/in_memory_game_repository.go
package repository

import (
	"sync"
	"tictactoe/internal/domain/entity"
	"tictactoe/internal/domain/port"
)

type inMemoryGameRepository struct {
	mu    sync.RWMutex
	games map[string]*entity.Game
}

func NewInMemoryGameRepository() port.GameRepository {
	return &inMemoryGameRepository{
		games: make(map[string]*entity.Game),
	}
}

func (r *inMemoryGameRepository) Save(game *entity.Game) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	// Deep copy to prevent external mutations
	gameCopy := *game
	boardCopy := make([][]string, len(game.Board))
	for i, row := range game.Board {
		boardCopy[i] = make([]string, len(row))
		copy(boardCopy[i], row)
	}
	gameCopy.Board = boardCopy
	
	r.games[game.ID] = &gameCopy
	return nil
}

func (r *inMemoryGameRepository) FindByID(id string) (*entity.Game, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	game, exists := r.games[id]
	if !exists {
		return nil, entity.ErrGameNotFound
	}
	
	// Deep copy to prevent external mutations
	gameCopy := *game
	boardCopy := make([][]string, len(game.Board))
	for i, row := range game.Board {
		boardCopy[i] = make([]string, len(row))
		copy(boardCopy[i], row)
	}
	gameCopy.Board = boardCopy
	
	return &gameCopy, nil
}

func (r *inMemoryGameRepository) FindPendingGames(boardSize, winningLength int) ([]*entity.Game, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var pendingGames []*entity.Game
	
	for _, game := range r.games {
		if game.Status == entity.StatusPending {
			// Filter by parameters if specified
			if boardSize > 0 && game.BoardSize != boardSize {
				continue
			}
			if winningLength > 0 && game.WinningLength != winningLength {
				continue
			}
			
			// Deep copy
			gameCopy := *game
			boardCopy := make([][]string, len(game.Board))
			for i, row := range game.Board {
				boardCopy[i] = make([]string, len(row))
				copy(boardCopy[i], row)
			}
			gameCopy.Board = boardCopy
			
			pendingGames = append(pendingGames, &gameCopy)
		}
	}
	
	return pendingGames, nil
}

func (r *inMemoryGameRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	delete(r.games, id)
	return nil
}

func (r *inMemoryGameRepository) Count() int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	return int64(len(r.games))
}
