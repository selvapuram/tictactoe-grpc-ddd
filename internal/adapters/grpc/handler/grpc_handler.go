// internal/adapters/grpc/handler/grpc_handler.go
package handler

import (
	"context"
	"tictactoe/internal/domain/entity"
	"tictactoe/internal/domain/port"
	pb "tictactoe/proto"
)

type GRPCHandler struct {
	pb.UnimplementedTicTacToeServiceServer
	gameService port.GameService
}

func NewGRPCHandler(gameService port.GameService) *GRPCHandler {
	return &GRPCHandler{
		gameService: gameService,
	}
}

func (h *GRPCHandler) StartGame(ctx context.Context, req *pb.StartGameRequest) (*pb.StartGameResponse, error) {
	boardSize := int(req.BoardSize)
	winningLength := int(req.WinningLength)
	
	game, err := h.gameService.StartGame(req.UserId, boardSize, winningLength)
	if err != nil {
		return &pb.StartGameResponse{
			Status:  mapGameStatusToProto(game.Status),
			Message: err.Error(),
		}, nil
	}

	var message string
	if game.Status == entity.StatusPending {
		message = "Game created. Waiting for opponent."
	} else {
		message = "Joined existing game. Game started!"
	}

	return &pb.StartGameResponse{
		GameId:  game.ID,
		Status:  mapGameStatusToProto(game.Status),
		Message: message,
	}, nil
}

func (h *GRPCHandler) SearchPendingGames(ctx context.Context, req *pb.SearchPendingGamesRequest) (*pb.SearchPendingGamesResponse, error) {
	boardSize := int(req.BoardSize)
	winningLength := int(req.WinningLength)
	
	games, err := h.gameService.SearchPendingGames(boardSize, winningLength)
	if err != nil {
		return nil, err
	}

	var pbGames []*pb.PendingGame
	for _, game := range games {
		pbGames = append(pbGames, &pb.PendingGame{
			GameId:        game.ID,
			CreatorId:     game.Player1ID,
			BoardSize:     int32(game.BoardSize),
			WinningLength: int32(game.WinningLength),
			CreatedAt:     game.CreatedAt.Unix(),
		})
	}

	return &pb.SearchPendingGamesResponse{
		Games: pbGames,
	}, nil
}

func (h *GRPCHandler) JoinGame(ctx context.Context, req *pb.JoinGameRequest) (*pb.JoinGameResponse, error) {
	game, err := h.gameService.JoinGame(req.UserId, req.GameId)
	if err != nil {
		return &pb.JoinGameResponse{
			Status:  pb.GameStatus_PENDING,
			Message: err.Error(),
		}, nil
	}

	return &pb.JoinGameResponse{
		Status:  mapGameStatusToProto(game.Status),
		Game:    mapGameToProto(game),
		Message: "Successfully joined game!",
	}, nil
}

func (h *GRPCHandler) MakeMove(ctx context.Context, req *pb.MakeMoveRequest) (*pb.MakeMoveResponse, error) {
	game, err := h.gameService.MakeMove(req.UserId, req.GameId, int(req.Row), int(req.Col))
	if err != nil {
		return &pb.MakeMoveResponse{
			Status:  pb.GameStatus_IN_PROGRESS,
			Message: err.Error(),
		}, nil
	}

	var message string
	switch game.Status {
	case entity.StatusFinishedWin:
		if game.WinnerID == req.UserId {
			message = "Congratulations! You won!"
		} else {
			message = "Game over. You lost."
		}
	case entity.StatusFinishedDraw:
		message = "Game over. It's a draw!"
	case entity.StatusInProgress:
		message = "Move successful. Waiting for opponent."
	}

	return &pb.MakeMoveResponse{
		Status:  mapGameStatusToProto(game.Status),
		Game:    mapGameToProto(game),
		Message: message,
	}, nil
}

func (h *GRPCHandler) GetGame(ctx context.Context, req *pb.GetGameRequest) (*pb.GetGameResponse, error) {
	game, err := h.gameService.GetGame(req.GameId, req.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.GetGameResponse{
		Game: mapGameToProto(game),
	}, nil
}

func (h *GRPCHandler) GetUserStats(ctx context.Context, req *pb.GetUserStatsRequest) (*pb.GetUserStatsResponse, error) {
	stats, err := h.gameService.GetUserStats(req.UserId)
	if err != nil {
		return nil, err
	}

	return &pb.GetUserStatsResponse{
		Stats: &pb.UserStats{
			UserId:     stats.UserID,
			Wins:       int32(stats.Wins),
			Losses:     int32(stats.Losses),
			Draws:      int32(stats.Draws),
			TotalGames: int32(stats.TotalGames),
		},
	}, nil
}

// Helper functions for mapping between domain and protobuf types

func mapGameStatusToProto(status entity.GameStatus) pb.GameStatus {
	switch status {
	case entity.StatusPending:
		return pb.GameStatus_PENDING
	case entity.StatusInProgress:
		return pb.GameStatus_IN_PROGRESS
	case entity.StatusFinishedWin:
		return pb.GameStatus_FINISHED_WIN
	case entity.StatusFinishedDraw:
		return pb.GameStatus_FINISHED_DRAW
	case entity.StatusAbandoned:
		return pb.GameStatus_ABANDONED
	default:
		return pb.GameStatus_PENDING
	}
}

func mapGameToProto(game *entity.Game) *pb.Game {
	return &pb.Game{
		Id:              game.ID,
		Player1Id:       game.Player1ID,
		Player2Id:       game.Player2ID,
		Board:           game.FlattenBoard(),
		BoardSize:       int32(game.BoardSize),
		WinningLength:   int32(game.WinningLength),
		Status:          mapGameStatusToProto(game.Status),
		CurrentPlayerId: game.CurrentPlayer,
		WinnerId:        game.WinnerID,
		CreatedAt:       game.CreatedAt.Unix(),
		UpdatedAt:       game.UpdatedAt.Unix(),
	}
}
