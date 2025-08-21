// internal/domain/entity/user_stats.go
package entity

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")
)

type UserStats struct {
	UserID     string
	Wins       int
	Losses     int
	Draws      int
	TotalGames int
}

func NewUserStats(userID string) *UserStats {
	return &UserStats{
		UserID: userID,
	}
}

func (s *UserStats) RecordWin() {
	s.Wins++
	s.TotalGames++
}

func (s *UserStats) RecordLoss() {
	s.Losses++
	s.TotalGames++
}

func (s *UserStats) RecordDraw() {
	s.Draws++
	s.TotalGames++
}
