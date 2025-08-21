// internal/application/service/game_service.go
package service

import (
	"tictactoe/internal/domain/config"
	"tictactoe/internal/domain/entity"
	"tictactoe/internal/domain/port"
)

type gameService struct {
	gameRepo port.GameRepository
	userRepo port.UserRepository
	config   *config.Config
}

func NewGameService(gameRepo port.GameRepository, userRepo port.UserRepository, cfg *config.Config) port.GameService {
	return &gameService{
		gameRepo: gameRepo,
		userRepo: userRepo,
		config:   cfg,
	}
}

func (s *gameService) StartGame(userID string, boardSize, winningLength int) (*entity.Game, error) {
	// Ensure user exists
	if err := s.userRepo.CreateUserIfNotExists(userID); err != nil {
		return nil, err
	}

	// Validate and normalize parameters
	boardSize = s.config.ValidateBoardSize(boardSize)
	winningLength = s.config.ValidateWinningLength(winningLength, boardSize)

	// Try to find an existing pending game with matching parameters
	pendingGames, err := s.gameRepo.FindPendingGames(boardSize, winningLength)
	if err != nil {
		return nil, err
	}

	// Look for a game that doesn't belong to the same user
	for _, game := range pendingGames {
		if game.Player1ID != userID {
			// Join this existing game
			if err := game.JoinPlayer(userID); err != nil {
				continue // Try next game
			}

			if err := s.gameRepo.Save(game); err != nil {
				return nil, err
			}

			return game, nil
		}
	}

	// No suitable pending game found, create a new one
	game := entity.NewGame(userID, boardSize, winningLength)
	if err := s.gameRepo.Save(game); err != nil {
		return nil, err
	}

	return game, nil
}

func (s *gameService) SearchPendingGames(boardSize, winningLength int) ([]*entity.Game, error) {
	return s.gameRepo.FindPendingGames(boardSize, winningLength)
}

func (s *gameService) JoinGame(userID, gameID string) (*entity.Game, error) {
	// Ensure user exists
	if err := s.userRepo.CreateUserIfNotExists(userID); err != nil {
		return nil, err
	}

	game, err := s.gameRepo.FindByID(gameID)
	if err != nil {
		return nil, err
	}

	if err := game.JoinPlayer(userID); err != nil {
		return nil, err
	}

	if err := s.gameRepo.Save(game); err != nil {
		return nil, err
	}

	return game, nil
}

func (s *gameService) MakeMove(userID, gameID string, row, col int) (*entity.Game, error) {
	game, err := s.gameRepo.FindByID(gameID)
	if err != nil {
		return nil, err
	}

	if !game.IsPlayerInGame(userID) {
		return nil, entity.ErrPlayerNotInGame
	}

	pos := entity.Position{Row: row, Col: col}
	if err := game.MakeMove(userID, pos); err != nil {
		return nil, err
	}

	// Update game
	if err := s.gameRepo.Save(game); err != nil {
		return nil, err
	}

	// Update user statistics if game is finished
	if game.Status == entity.StatusFinishedWin || game.Status == entity.StatusFinishedDraw {
		if err := s.updateUserStats(game); err != nil {
			// Log error but don't fail the move
			// In production, this would be logged properly
		}
	}

	return game, nil
}

func (s *gameService) GetGame(gameID, userID string) (*entity.Game, error) {
	game, err := s.gameRepo.FindByID(gameID)
	if err != nil {
		return nil, err
	}

	// Ensure user is part of the game (privacy)
	if !game.IsPlayerInGame(userID) {
		return nil, entity.ErrPlayerNotInGame
	}

	return game, nil
}

func (s *gameService) GetUserStats(userID string) (*entity.UserStats, error) {
	stats, err := s.userRepo.FindStatsByUserID(userID)
	if err != nil {
		// If user doesn't exist, create new stats
		if err := s.userRepo.CreateUserIfNotExists(userID); err != nil {
			return nil, err
		}
		stats = entity.NewUserStats(userID)
	}

	return stats, nil
}

func (s *gameService) updateUserStats(game *entity.Game) error {
	// Get or create stats for both players
	player1Stats, err := s.userRepo.FindStatsByUserID(game.Player1ID)
	if err != nil {
		player1Stats = entity.NewUserStats(game.Player1ID)
	}

	player2Stats, err := s.userRepo.FindStatsByUserID(game.Player2ID)
	if err != nil {
		player2Stats = entity.NewUserStats(game.Player2ID)
	}

	// Update based on game outcome
	switch game.Status {
	case entity.StatusFinishedWin:
		if game.WinnerID == game.Player1ID {
			player1Stats.RecordWin()
			player2Stats.RecordLoss()
		} else {
			player2Stats.RecordWin()
			player1Stats.RecordLoss()
		}
	case entity.StatusFinishedDraw:
		player1Stats.RecordDraw()
		player2Stats.RecordDraw()
	}

	// Save updated stats
	if err := s.userRepo.SaveStats(player1Stats); err != nil {
		return err
	}
	if err := s.userRepo.SaveStats(player2Stats); err != nil {
		return err
	}

	return nil
}
