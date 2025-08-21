// internal/domain/config/config.go
package config

type Config struct {
	DefaultBoardSize     int
	DefaultWinningLength int
	MaxBoardSize         int
	MinBoardSize         int
}

func DefaultConfig() *Config {
	return &Config{
		DefaultBoardSize:     3,
		DefaultWinningLength: 3,
		MaxBoardSize:         20, // Reasonable limit for scalability
		MinBoardSize:         3,
	}
}

func (c *Config) ValidateBoardSize(size int) int {
	if size < c.MinBoardSize {
		return c.DefaultBoardSize
	}
	if size > c.MaxBoardSize {
		return c.MaxBoardSize
	}
	return size
}

func (c *Config) ValidateWinningLength(length, boardSize int) int {
	if length <= 0 {
		return boardSize // Default to board size
	}
	if length > boardSize {
		return boardSize
	}
	return length
}
