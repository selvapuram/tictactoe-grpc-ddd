// test/integration/game_integration_test.go
package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"tictactoe/internal/adapters/grpc/handler"
	"tictactoe/internal/adapters/repository"
	"tictactoe/internal/application/service"
	"tictactoe/internal/domain/config"
	pb "tictactoe/proto"
)

func setupTestServer() *handler.GRPCHandler {
	gameRepo := repository.NewInMemoryGameRepository()
	userRepo := repository.NewInMemoryUserRepository()
	cfg := config.DefaultConfig()
	gameService := service.NewGameService(gameRepo, userRepo, cfg)
	return handler.NewGRPCHandler(gameService)
}

func TestCompleteGameFlow(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	// Player1 starts a game
	startResp, err := server.StartGame(ctx, &pb.StartGameRequest{
		UserId:        "player1",
		BoardSize:     3,
		WinningLength: 3,
	})
	require.NoError(t, err)
	assert.Equal(t, pb.GameStatus_PENDING, startResp.Status)
	gameID := startResp.GameId

	// Search for pending games
	searchResp, err := server.SearchPendingGames(ctx, &pb.SearchPendingGamesRequest{})
	require.NoError(t, err)
	assert.Len(t, searchResp.Games, 1)
	assert.Equal(t, gameID, searchResp.Games[0].GameId)

	// Player2 joins the game
	joinResp, err := server.JoinGame(ctx, &pb.JoinGameRequest{
		UserId: "player2",
		GameId: gameID,
	})
	require.NoError(t, err)
	assert.Equal(t, pb.GameStatus_IN_PROGRESS, joinResp.Status)

	// Play the game - Player1 wins
	moves := []struct {
		player string
		row    int32
		col    int32
	}{
		{"player1", 0, 0}, // X
		{"player2", 1, 0}, // O
		{"player1", 0, 1}, // X
		{"player2", 1, 1}, // O
		{"player1", 0, 2}, // X wins
	}

	var lastResp *pb.MakeMoveResponse
	for _, move := range moves {
		moveResp, err := server.MakeMove(ctx, &pb.MakeMoveRequest{
			UserId: move.player,
			GameId: gameID,
			Row:    move.row,
			Col:    move.col,
		})
		require.NoError(t, err)
		lastResp = moveResp
	}

	// Verify game is won
	assert.Equal(t, pb.GameStatus_FINISHED_WIN, lastResp.Status)
	assert.Equal(t, "player1", lastResp.Game.WinnerId)

	// Check final game state
	gameResp, err := server.GetGame(ctx, &pb.GetGameRequest{
		GameId: gameID,
		UserId: "player1",
	})
	require.NoError(t, err)

	expectedBoard := []string{"X", "X", "X", "O", "O", "", "", "", ""}
	assert.Equal(t, expectedBoard, gameResp.Game.Board)

	// Check user statistics
	stats1, err := server.GetUserStats(ctx, &pb.GetUserStatsRequest{UserId: "player1"})
	require.NoError(t, err)
	assert.Equal(t, int32(1), stats1.Stats.Wins)
	assert.Equal(t, int32(0), stats1.Stats.Losses)
	assert.Equal(t, int32(1), stats1.Stats.TotalGames)

	stats2, err := server.GetUserStats(ctx, &pb.GetUserStatsRequest{UserId: "player2"})
	require.NoError(t, err)
	assert.Equal(t, int32(0), stats2.Stats.Wins)
	assert.Equal(t, int32(1), stats2.Stats.Losses)
	assert.Equal(t, int32(1), stats2.Stats.TotalGames)
}

func TestDrawGame(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	// Setup game
	startResp, _ := server.StartGame(ctx, &pb.StartGameRequest{
		UserId:        "player1",
		BoardSize:     3,
		WinningLength: 3,
	})
	gameID := startResp.GameId

	server.JoinGame(ctx, &pb.JoinGameRequest{
		UserId: "player2",
		GameId: gameID,
	})

	// Play to a draw
	moves := []struct {
		player string
		row    int32
		col    int32
	}{
		{"player1", 0, 0}, // X
		{"player2", 0, 1}, // O
		{"player1", 0, 2}, // X
		{"player2", 1, 0}, // O
		{"player1", 1, 2}, // X
		{"player2", 1, 1}, // O
		{"player1", 2, 0}, // X
		{"player2", 2, 1}, // O
		{"player1", 2, 2}, // X
	}

	var lastResp *pb.MakeMoveResponse
	for _, move := range moves {
		moveResp, _ := server.MakeMove(ctx, &pb.MakeMoveRequest{
			UserId: move.player,
			GameId: gameID,
			Row:    move.row,
			Col:    move.col,
		})
		lastResp = moveResp
	}

	// Verify draw
	assert.Equal(t, pb.GameStatus_FINISHED_DRAW, lastResp.Status)
	assert.Empty(t, lastResp.Game.WinnerId)

	// Check statistics
	stats1, _ := server.GetUserStats(ctx, &pb.GetUserStatsRequest{UserId: "player1"})
	stats2, _ := server.GetUserStats(ctx, &pb.GetUserStatsRequest{UserId: "player2"})

	assert.Equal(t, int32(1), stats1.Stats.Draws)
	assert.Equal(t, int32(1), stats2.Stats.Draws)
}

func TestErrorConditions(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	// Test invalid move on non-existent game
	moveResp, err := server.MakeMove(ctx, &pb.MakeMoveRequest{
		UserId: "player1",
		GameId: "nonexistent",
		Row:    0,
		Col:    0,
	})
	require.NoError(t, err) // gRPC doesn't return error for business logic errors
	assert.Contains(t, moveResp.Message, "game not found")

	// Test joining non-existent game
	joinResp, err := server.JoinGame(ctx, &pb.JoinGameRequest{
		UserId: "player1",
		GameId: "nonexistent",
	})
	require.NoError(t, err)
	assert.Contains(t, joinResp.Message, "game not found")
}

func TestConcurrentAccess(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	// Start game
	startResp, _ := server.StartGame(ctx, &pb.StartGameRequest{
		UserId:        "player1",
		BoardSize:     3,
		WinningLength: 3,
	})
	gameID := startResp.GameId

	server.JoinGame(ctx, &pb.JoinGameRequest{
		UserId: "player2",
		GameId: gameID,
	})

	// Test concurrent moves (should handle race conditions)
	done := make(chan bool, 2)

	go func() {
		server.MakeMove(ctx, &pb.MakeMoveRequest{
			UserId: "player1",
			GameId: gameID,
			Row:    0,
			Col:    0,
		})
		done <- true
	}()

	go func() {
		server.MakeMove(ctx, &pb.MakeMoveRequest{
			UserId: "player2",
			GameId: gameID,
			Row:    0,
			Col:    1,
		})
		done <- true
	}()

	// Wait for both goroutines
	<-done
	<-done

	// Verify game state is consistent
	gameResp, _ := server.GetGame(ctx, &pb.GetGameRequest{
		GameId: gameID,
		UserId: "player1",
	})

	// At least one move should have been made
	moveCount := 0
	for _, cell := range gameResp.Game.Board {
		if cell != "" {
			moveCount++
		}
	}
	assert.True(t, moveCount >= 1)
}

func TestCustomBoardSize(t *testing.T) {
	server := setupTestServer()
	ctx := context.Background()

	// Test 5x5 board with winning length 4
	startResp, err := server.StartGame(ctx, &pb.StartGameRequest{
		UserId:        "player1",
		BoardSize:     5,
		WinningLength: 4,
	})
	require.NoError(t, err)
	gameID := startResp.GameId

	server.JoinGame(ctx, &pb.JoinGameRequest{
		UserId: "player2",
		GameId: gameID,
	})

	gameResp, _ := server.GetGame(ctx, &pb.GetGameRequest{
		GameId: gameID,
		UserId: "player1",
	})

	assert.Equal(t, int32(5), gameResp.Game.BoardSize)
	assert.Equal(t, int32(4), gameResp.Game.WinningLength)
	assert.Len(t, gameResp.Game.Board, 25) // 5x5 = 25 cells
}
