// internal/application/service/game_service_test.go
package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tictactoe/internal/adapters/repository"
	"tictactoe/internal/domain/config"
	"tictactoe/internal/domain/entity"
)

func TestGameService_StartGame(t *testing.T) {
	gameRepo := repository.NewInMemoryGameRepository()
	userRepo := repository.NewInMemoryUserRepository()
	cfg := config.DefaultConfig()
	service := NewGameService(gameRepo, userRepo, cfg)

	// Test creating new game
	game, err := service.StartGame("player1", 3, 3)
	require.NoError(t, err)
	assert.Equal(t, entity.StatusPending, game.Status)
	assert.Equal(t, "player1", game.Player1ID)

	// Test joining existing game
	game2, err := service.StartGame("player2", 3, 3)
	require.NoError(t, err)
	assert.Equal(t, entity.StatusInProgress, game2.Status)
	assert.Equal(t, game.ID, game2.ID) // Same game
	assert.Equal(t, "player2", game2.Player2ID)
}

func TestGameService_MakeMove(t *testing.T) {
	gameRepo := repository.NewInMemoryGameRepository()
	userRepo := repository.NewInMemoryUserRepository()
	cfg := config.DefaultConfig()
	service := NewGameService(gameRepo, userRepo, cfg)

	// Setup game
	game, _ := service.StartGame("player1", 3, 3)
	game, _ = service.JoinGame("player2", game.ID)

	// Test valid move - player 1
	game, err := service.MakeMove("player1", game.ID, 0, 0)
	require.NoError(t, err)
	assert.Equal(t, "X", game.Board[0][0])

	// Test invalid move
	_, err = service.MakeMove("player1", game.ID, 0, 0)
	assert.Equal(t, entity.ErrNotPlayersTurn, err)

	// Test valid move - player 2
	game, err = service.MakeMove("player2", game.ID, 1, 0)
	require.NoError(t, err)
	assert.Equal(t, "O", game.Board[1][0])

	// Test Invalid move - player 1
	_, err = service.MakeMove("player1", game.ID, 0, 0)
	assert.Equal(t, entity.ErrPositionOccupied, err)
}

func TestGameService_GetUserStats(t *testing.T) {
	gameRepo := repository.NewInMemoryGameRepository()
	userRepo := repository.NewInMemoryUserRepository()
	cfg := config.DefaultConfig()
	service := NewGameService(gameRepo, userRepo, cfg)

	// Test getting stats for new user
	stats, err := service.GetUserStats("newuser")
	require.NoError(t, err)
	assert.Equal(t, "newuser", stats.UserID)
	assert.Equal(t, 0, stats.TotalGames)
}

func TestGameService_CompleteGame(t *testing.T) {
	gameRepo := repository.NewInMemoryGameRepository()
	userRepo := repository.NewInMemoryUserRepository()
	cfg := config.DefaultConfig()
	service := NewGameService(gameRepo, userRepo, cfg)

	// Setup and complete a game
	game, _ := service.StartGame("player1", 3, 3)
	game, _ = service.JoinGame("player2", game.ID)

	// Player1 wins
	service.MakeMove("player1", game.ID, 0, 0)           // X
	service.MakeMove("player2", game.ID, 1, 0)           // O
	service.MakeMove("player1", game.ID, 0, 1)           // X
	service.MakeMove("player2", game.ID, 1, 1)           // O
	game, _ = service.MakeMove("player1", game.ID, 0, 2) // X wins

	assert.Equal(t, entity.StatusFinishedWin, game.Status)
	assert.Equal(t, "player1", game.WinnerID)

	// Check stats
	stats1, _ := service.GetUserStats("player1")
	stats2, _ := service.GetUserStats("player2")

	assert.Equal(t, 1, stats1.Wins)
	assert.Equal(t, 0, stats1.Losses)
	assert.Equal(t, 1, stats1.TotalGames)

	assert.Equal(t, 0, stats2.Wins)
	assert.Equal(t, 1, stats2.Losses)
	assert.Equal(t, 1, stats2.TotalGames)
}
