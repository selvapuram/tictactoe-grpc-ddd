// internal/domain/entity/game_test.go
package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewGame(t *testing.T) {
	tests := []struct {
		name          string
		boardSize     int
		winningLength int
		expectedBS    int
		expectedWL    int
	}{
		{"default values", 0, 0, 3, 3},
		{"custom values", 5, 4, 5, 4},
		{"winning length too large", 3, 5, 3, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			game := NewGame("player1", tt.boardSize, tt.winningLength)

			assert.Equal(t, tt.expectedBS, game.BoardSize)
			assert.Equal(t, tt.expectedWL, game.WinningLength)
			assert.Equal(t, StatusPending, game.Status)
			assert.Equal(t, "player1", game.Player1ID)
			assert.Empty(t, game.Player2ID)
		})
	}
}

func TestGame_JoinPlayer(t *testing.T) {
	game := NewGame("player1", 3, 3)

	// Test successful join
	err := game.JoinPlayer("player2")
	assert.NoError(t, err)
	assert.Equal(t, "player2", game.Player2ID)
	assert.Equal(t, StatusInProgress, game.Status)
	assert.Equal(t, "player1", game.CurrentPlayer)

	// Test joining full game
	err = game.JoinPlayer("player3")
	assert.Equal(t, ErrGameFull, err)

	// Test same player joining twice
	game2 := NewGame("player1", 3, 3)
	err = game2.JoinPlayer("player1")
	assert.Equal(t, ErrGameFull, err)
}

func TestGame_MakeMove(t *testing.T) {
	game := NewGame("player1", 3, 3)
	game.JoinPlayer("player2")

	// Test valid move
	err := game.MakeMove("player1", Position{0, 0})
	assert.NoError(t, err)
	assert.Equal(t, "X", game.Board[0][0])
	assert.Equal(t, "player2", game.CurrentPlayer)

	// Test move out of turn
	err = game.MakeMove("player1", Position{0, 1})
	assert.Equal(t, ErrNotPlayersTurn, err)

	// Test invalid position
	err = game.MakeMove("player2", Position{5, 5})
	assert.Equal(t, ErrInvalidMove, err)

	// Test occupied position
	err = game.MakeMove("player2", Position{0, 0})
	assert.Equal(t, ErrPositionOccupied, err)

	// Test valid move by player2
	err = game.MakeMove("player2", Position{0, 1})
	assert.NoError(t, err)
	assert.Equal(t, "O", game.Board[0][1])
	assert.Equal(t, "player1", game.CurrentPlayer)
}

func TestGame_WinConditions(t *testing.T) {
	game := NewGame("player1", 3, 3)
	game.JoinPlayer("player2")

	// Create winning condition - horizontal
	game.MakeMove("player1", Position{0, 0}) // X
	game.MakeMove("player2", Position{1, 0}) // O
	game.MakeMove("player1", Position{0, 1}) // X
	game.MakeMove("player2", Position{1, 1}) // O
	game.MakeMove("player1", Position{0, 2}) // X - wins

	assert.Equal(t, StatusFinishedWin, game.Status)
	assert.Equal(t, "player1", game.WinnerID)
}

func TestGame_DrawCondition(t *testing.T) {
	game := NewGame("player1", 3, 3)
	game.JoinPlayer("player2")

	// Fill board without winner
	moves := []Position{
		{0, 0}, {0, 1}, {0, 2}, // X O X
		{1, 1}, {1, 0}, {1, 2}, // O X O
		{2, 1}, {2, 0}, {2, 2}, // O X X
	}

	for i, move := range moves {
		player := "player1"
		if i%2 == 1 {
			player = "player2"
		}
		game.MakeMove(player, move)
	}

	assert.Equal(t, StatusFinishedDraw, game.Status)
}

func TestPosition_IsValid(t *testing.T) {
	tests := []struct {
		pos       Position
		boardSize int
		expected  bool
	}{
		{Position{0, 0}, 3, true},
		{Position{2, 2}, 3, true},
		{Position{-1, 0}, 3, false},
		{Position{0, 3}, 3, false},
		{Position{3, 0}, 3, false},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.pos.IsValid(tt.boardSize))
	}
}
