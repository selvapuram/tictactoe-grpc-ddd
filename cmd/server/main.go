// cmd/server/main.go
package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"tictactoe/internal/adapters/grpc/handler"
	"tictactoe/internal/adapters/repository"
	"tictactoe/internal/application/service"
	"tictactoe/internal/domain/config"
	pb "tictactoe/proto"
)

func main() {
	// Load configuration
	cfg := config.DefaultConfig()

	// Initialize repositories (in-memory)
	gameRepo := repository.NewInMemoryGameRepository()
	userRepo := repository.NewInMemoryUserRepository()

	// Initialize services
	gameService := service.NewGameService(gameRepo, userRepo, cfg)

	// Initialize gRPC handler
	grpcHandler := handler.NewGRPCHandler(gameService)

	// Setup gRPC server
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	server := grpc.NewServer()
	pb.RegisterTicTacToeServiceServer(server, grpcHandler)

	// Enable reflection for testing
	reflection.Register(server)

	// Graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		log.Println("Starting gRPC server on :8080")
		if err := server.Serve(lis); err != nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 10*time.Second)
	defer shutdownCancel()

	done := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(done)
	}()

	select {
	case <-shutdownCtx.Done():
		log.Println("Shutdown timeout exceeded, forcing stop")
		server.Stop()
	case <-done:
		log.Println("Server stopped gracefully")
	}
}
