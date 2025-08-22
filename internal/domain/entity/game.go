// internal/domain/entity/game.go
package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrGameNotFound     = errors.New("game not found")
	ErrGameFull         = errors.New("game is full")
	ErrNotPlayersTurn   = errors.New("not player's turn")
	ErrGameFinished     = errors.New("game is finished")
	ErrInvalidMove      = errors.New("invalid move")
	ErrPositionOccupied = errors.New("position already occupied")
	ErrPlayerNotInGame  = errors.New("player not in game")
)

type GameStatus int

const (
	StatusPending GameStatus = iota
	StatusInProgress
	StatusFinishedWin
	StatusFinishedDraw
	StatusAbandoned
)

type Position struct {
	Row int
	Col int
}

func (p Position) IsValid(boardSize int) bool {
	return p.Row >= 0 && p.Row < boardSize && p.Col >= 0 && p.Col < boardSize
}

type Game struct {
	ID            string
	Player1ID     string
	Player2ID     string
	Board         [][]string
	BoardSize     int
	WinningLength int
	Status        GameStatus
	CurrentPlayer string
	WinnerID      string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewGame(player1ID string, boardSize, winningLength int) *Game {
	if boardSize <= 0 {
		boardSize = 3
	}
	if winningLength <= 0 || winningLength > boardSize {
		winningLength = boardSize
	}

	board := make([][]string, boardSize)
	for i := range board {
		board[i] = make([]string, boardSize)
	}

	return &Game{
		ID:            uuid.New().String(),
		Player1ID:     player1ID,
		Board:         board,
		BoardSize:     boardSize,
		WinningLength: winningLength,
		Status:        StatusPending,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

func (g *Game) JoinPlayer(playerID string) error {
	if g.Status != StatusPending {
		return ErrGameFull
	}

	if g.Player1ID == playerID {
		return ErrGameFull // Same player cannot join twice
	}

	g.Player2ID = playerID
	g.CurrentPlayer = g.Player1ID // Player 1 always starts
	g.Status = StatusInProgress
	g.UpdatedAt = time.Now()
	return nil
}

func (g *Game) MakeMove(playerID string, pos Position) error {
	err := g.validate(playerID, pos)
	if err != nil {
		return err
	}

	// Determine player symbol
	symbol := g.GetPlayerSymbol(playerID)

	// Make the move
	g.Board[pos.Row][pos.Col] = symbol
	g.UpdatedAt = time.Now()

	// Check for winner
	if g.checkWinner(pos, symbol) {
		g.setToWin(playerID)
	} else if g.isBoardFull() {
		g.setToDraw()
	} else {
		g.switchTurn()
	}

	return nil
}

func (g *Game) setToWin(playerID string) {
	g.Status = StatusFinishedWin
	g.WinnerID = playerID
}

func (g *Game) setToDraw() {
	g.Status = StatusFinishedDraw
}

func (g *Game) switchTurn() {
	if g.CurrentPlayer == g.Player1ID {
		g.CurrentPlayer = g.Player2ID
	} else {
		g.CurrentPlayer = g.Player1ID
	}
}

func (g *Game) validate(playerID string, pos Position) error {
	if g.Status != StatusInProgress {
		return ErrGameFinished
	}

	if g.CurrentPlayer != playerID {
		return ErrNotPlayersTurn
	}

	if !pos.IsValid(g.BoardSize) {
		return ErrInvalidMove
	}

	if g.Board[pos.Row][pos.Col] != "" {
		return ErrPositionOccupied
	}
	return nil
}

func (g *Game) IsPlayerInGame(playerID string) bool {
	return g.Player1ID == playerID || g.Player2ID == playerID
}

func (g *Game) GetPlayerSymbol(playerID string) string {
	if playerID == g.Player1ID {
		return "X"
	}
	if playerID == g.Player2ID {
		return "O"
	}
	return ""
}

func (g *Game) checkWinner(lastMove Position, symbol string) bool {
	// Check all four directions: horizontal, vertical, diagonal, anti-diagonal
	directions := [][2]int{
		{0, 1},  // horizontal
		{1, 0},  // vertical
		{1, 1},  // diagonal
		{1, -1}, // anti-diagonal
	}

	for _, dir := range directions {
		if g.countInDirection(lastMove, symbol, dir[0], dir[1]) >= g.WinningLength {
			return true
		}
	}
	return false
}

func (g *Game) countInDirection(pos Position, symbol string, deltaRow, deltaCol int) int {
	count := 1 // Count the current position

	// Count in positive direction
	r, c := pos.Row+deltaRow, pos.Col+deltaCol
	for r >= 0 && r < g.BoardSize && c >= 0 && c < g.BoardSize && g.Board[r][c] == symbol {
		count++
		r += deltaRow
		c += deltaCol
	}

	// Count in negative direction
	r, c = pos.Row-deltaRow, pos.Col-deltaCol
	for r >= 0 && r < g.BoardSize && c >= 0 && c < g.BoardSize && g.Board[r][c] == symbol {
		count++
		r -= deltaRow
		c -= deltaCol
	}

	return count
}

func (g *Game) isBoardFull() bool {
	for i := 0; i < g.BoardSize; i++ {
		for j := 0; j < g.BoardSize; j++ {
			if g.Board[i][j] == "" {
				return false
			}
		}
	}
	return true
}

func (g *Game) FlattenBoard() []string {
	flat := make([]string, g.BoardSize*g.BoardSize)
	for i := 0; i < g.BoardSize; i++ {
		for j := 0; j < g.BoardSize; j++ {
			flat[i*g.BoardSize+j] = g.Board[i][j]
		}
	}
	return flat
}
