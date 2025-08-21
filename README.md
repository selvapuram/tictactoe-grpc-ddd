# Tic-Tac-Toe Backend Service

A production-grade, scalable backend service for multiplayer tic-tac-toe games built with Go, gRPC, and hexagonal architecture.

## Features

- **Real-time multiplayer**: Two players can play simultaneously
- **Configurable board**: Customizable board size and winning length
- **Game matchmaking**: Automatic pairing of players or manual game joining
- **Statistics tracking**: Win/loss/draw statistics per user
- **Production-ready**: Comprehensive testing, logging, and error handling
- **Scalable architecture**: Designed for millions of users with proper separation of concerns

## Architecture

The service follows **Domain-Driven Design (DDD)** and **Hexagonal Architecture** principles:

```
├── cmd/server/          # Application entry point
├── internal/
│   ├── domain/          # Core business logic
│   │   ├── entity/      # Domain entities (Game, UserStats)
│   │   ├── port/        # Interfaces (repositories, services)
│   │   └── config/      # Domain configuration
│   ├── application/     # Application services
│   │   └── service/     # Business use cases
│   └── adapters/        # External adapters
│       ├── grpc/        # gRPC handlers
│       └── repository/  # Data persistence
├── proto/               # Protocol buffer definitions
└── test/               # Integration tests
```

### Key Design Decisions

1. **In-Memory Storage**: As requested, all data is stored in memory using concurrent-safe maps with mutex protection
2. **gRPC API**: Provides type-safe, high-performance communication
3. **Hexagonal Architecture**: Ensures clean separation between business logic and external concerns
4. **Configurable Game Rules**: Board size and winning length can be customized per game
5. **Automatic Matchmaking**: Players are automatically paired when starting games with matching parameters

## API Documentation

### gRPC Service Definition

```protobuf
service TicTacToeService {
  rpc StartGame(StartGameRequest) returns (StartGameResponse);
  rpc SearchPendingGames(SearchPendingGamesRequest) returns (SearchPendingGamesResponse);
  rpc JoinGame(JoinGameRequest) returns (JoinGameResponse);
  rpc MakeMove(MakeMoveRequest) returns (MakeMoveResponse);
  rpc GetGame(GetGameRequest) returns (GetGameResponse);
  rpc GetUserStats(GetUserStatsRequest) returns (GetUserStatsResponse);
}
```

### Example Usage

1. **Start a Game**:
   ```
   StartGame(user_id="player1", board_size=3, winning_length=3)
   → Returns game_id and status (PENDING if waiting, IN_PROGRESS if joined existing game)
   ```

2. **Join a Game**:
   ```
   JoinGame(user_id="player2", game_id="uuid")
   → Game status changes to IN_PROGRESS
   ```

3. **Make a Move**:
   ```
   MakeMove(user_id="player1", game_id="uuid", row=0, col=0)
   → Places X or O, checks for win/draw, switches turns
   ```

## Building and Running

### Prerequisites

- Go 1.21+
- Protocol Buffers compiler (`protoc`)
- Make (optional, for convenience)

### Quick Start

```bash
# Clone the repository
git clone <repository-url>
cd tictactoe

# Install development tools and dependencies
make dev

# Build and run
make run
```

The server will start on port 8080.

### Manual Build

```bash
# Download dependencies
go mod download

# Generate protobuf code
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/*.proto

# Build
go build -o tictactoe-server ./cmd/server

# Run
./tictactoe-server
```

### Docker

```bash
# Build and run with Docker
make docker-run

# Or manually:
docker build -t tictactoe .
docker run -p 8080:8080 tictactoe
```

### Testing

```bash
# Run all tests
make test

# Run only unit tests
make test-unit

# Run only integration tests
make test-integration

# Generate coverage report
make test-coverage
```

## Testing the API

You can test the gRPC API using tools like:

1. **grpcurl** (recommended):
   ```bash
   # List services
   grpcurl -plaintext localhost:8080 list

   # Start a game
   grpcurl -plaintext -d '{"user_id":"player1","board_size":3,"winning_length":3}' \
     localhost:8080 tictactoe.TicTacToeService/StartGame
   ```

2. **BloomRPC** or **Postman** with gRPC support

3. **Custom Go client**:
   ```go
   conn, _ := grpc.Dial("localhost:8080", grpc.WithInsecure())
   client := pb.NewTicTacToeServiceClient(conn)
   response, _ := client.StartGame(context.Background(), &pb.StartGameRequest{
       UserId: "player1",
       BoardSize: 3,
       WinningLength: 3,
   })
   ```

## Scalability Considerations

### Current Architecture Benefits

1. **Stateless Design**: Each request is self-contained, enabling horizontal scaling
2. **In-Memory Storage**: Ultra-fast access times, suitable for real-time gaming
3. **Concurrent Safety**: All repositories use proper locking mechanisms
4. **Resource Efficiency**: Minimal memory footprint per game (~1KB)

### Production Scaling Strategies

1. **Horizontal Scaling**:
   - Deploy multiple instances behind a load balancer
   - Use consistent hashing for game distribution
   - Implement sticky sessions for active games

2. **Memory Management**:
   - Implement game cleanup for abandoned games
   - Add TTL for inactive games
   - Monitor memory usage with metrics

3. **Database Migration**:
   - Replace in-memory storage with Redis/distributed cache
   - Add persistent storage for long-term statistics
   - Implement data partitioning strategies

4. **Performance Optimizations**:
   - Connection pooling
   - Request batching for statistics
   - Caching frequently accessed data

### Estimated Capacity

With current in-memory design:
- **Memory per game**: ~1KB
- **Memory per user stats**: ~100 bytes
- **Concurrent games**: ~100K games per GB RAM
- **Total users**: Millions (with periodic cleanup)

## Trade-offs Made

### 1. In-Memory vs. Persistent Storage
**Decision**: In-memory storage as requested
**Trade-offs**:
- ✅ Ultra-fast performance, no I/O latency
- ✅ Simple implementation
- ❌ Data lost on restart
- ❌ Limited by server memory
- ❌ No data durability

### 2. gRPC vs. REST
**Decision**: gRPC (with REST fallback mentioned)
**Trade-offs**:
- ✅ Type safety and performance
- ✅ Bi-directional streaming capability
- ✅ Automatic client generation
- ❌ Less browser-friendly
- ❌ Steeper learning curve

### 3. Automatic Matchmaking vs. Manual Only
**Decision**: Automatic matchmaking with manual option
**Trade-offs**:
- ✅ Better user experience
- ✅ Faster game starts
- ❌ Less control over opponent selection
- ❌ More complex matching logic

### 4. Synchronous vs. Asynchronous Processing
**Decision**: Synchronous request handling
**Trade-offs**:
- ✅ Simpler error handling and debugging
- ✅ Immediate feedback to users
- ❌ May block on high load
- ❌ Less throughput under extreme load

### 5. Single Service vs. Microservices
**Decision**: Single service (monolith)
**Trade-offs**:
- ✅ Simpler deployment and debugging
- ✅ Better performance (no network calls)
- ✅ Easier development for small team
- ❌ Less flexibility for independent scaling
- ❌ Single point of failure

## Future Enhancements

1. **Real-time Notifications**: WebSocket/Server-Sent Events for live updates
2. **Tournament Mode**: Multi-player tournaments with brackets
3. **AI Opponents**: Computer players with different difficulty levels
4. **Game Replay**: Store and replay completed games
5. **Leaderboards**: Global and time-based rankings
6. **Authentication**: Proper user management and security
7. **Monitoring**: Metrics, logging, and health checks
8. **Admin API**: Game management and user administration

## Contributing

1. Follow the existing code style and architecture patterns
2. Add tests for new features
3. Update documentation for API changes
4. Run `make lint` before submitting changes

## License

[Add your license here]