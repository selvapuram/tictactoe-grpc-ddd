# scripts/test-api.sh
#!/bin/bash
# Simple API testing script using grpcurl

set -e

SERVER_ADDR="localhost:8080"

echo "Testing Tic-Tac-Toe API..."

# Check if grpcurl is available
if ! command -v grpcurl &> /dev/null; then
    echo "grpcurl not found. Install it with: go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest"
    exit 1
fi

echo "1. Testing server connection..."
grpcurl -plaintext $SERVER_ADDR list tictactoe.TicTacToeService || {
    echo "Server not responding. Make sure it's running on $SERVER_ADDR"
    exit 1
}

echo "✓ Server is running"

echo "2. Starting a game..."
GAME_RESPONSE=$(grpcurl -plaintext -d '{"user_id":"player1","board_size":3,"winning_length":3}' \
    $SERVER_ADDR tictactoe.TicTacToeService/StartGame)
echo "Response: $GAME_RESPONSE"

# Extract game ID (simple regex - in production use proper JSON parsing)
GAME_ID=$(echo $GAME_RESPONSE | grep -o '"game_id":"[^"]*"' | cut -d'"' -f4)
echo "Game ID: $GAME_ID"

echo "3. Searching pending games..."
grpcurl -plaintext -d '{}' $SERVER_ADDR tictactoe.TicTacToeService/SearchPendingGames

echo "4. Player 2 joins the game..."
grpcurl -plaintext -d "{\"user_id\":\"player2\",\"game_id\":\"$GAME_ID\"}" \
    $SERVER_ADDR tictactoe.TicTacToeService/JoinGame

echo "5. Player 1 makes a move..."
grpcurl -plaintext -d "{\"user_id\":\"player1\",\"game_id\":\"$GAME_ID\",\"row\":0,\"col\":0}" \
    $SERVER_ADDR tictactoe.TicTacToeService/MakeMove

echo "6. Getting game state..."
grpcurl -plaintext -d "{\"user_id\":\"player1\",\"game_id\":\"$GAME_ID\"}" \
    $SERVER_ADDR tictactoe.TicTacToeService/GetGame

echo "7. Getting user stats..."
grpcurl -plaintext -d '{"user_id":"player1"}' \
    $SERVER_ADDR tictactoe.TicTacToeService/GetUserStats

echo "✓ All tests completed successfully!"
