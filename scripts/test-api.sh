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
GAME_ID=$(echo $GAME_RESPONSE | grep -o '"gameId": "[^"]*"' | cut -d':' -f2 | tr -d '"' | xargs)
echo "Game ID:$GAME_ID"

echo "3. Searching pending games..."
grpcurl -plaintext -d '{}' $SERVER_ADDR tictactoe.TicTacToeService/SearchPendingGames

echo "4. Player 2 joins the game..."
grpcurl -plaintext -d "{\"user_id\":\"player2\",\"game_id\":\"$GAME_ID\"}" \
    $SERVER_ADDR tictactoe.TicTacToeService/JoinGame

echo "5. Player 1 makes a move..."
grpcurl -plaintext -d "{\"user_id\":\"player1\",\"game_id\":\"$GAME_ID\",\"row\":0,\"col\":0}" \
    $SERVER_ADDR tictactoe.TicTacToeService/MakeMove

echo "6. Player 2 makes a move..."
grpcurl -plaintext -d "{\"user_id\":\"player2\",\"game_id\":\"$GAME_ID\",\"row\":0,\"col\":1}" \
    $SERVER_ADDR tictactoe.TicTacToeService/MakeMove

echo "7. Player 1 makes a move..."
grpcurl -plaintext -d "{\"user_id\":\"player1\",\"game_id\":\"$GAME_ID\",\"row\":1,\"col\":1}" \
    $SERVER_ADDR tictactoe.TicTacToeService/MakeMove

echo "8. Player 2 makes a move..."
grpcurl -plaintext -d "{\"user_id\":\"player2\",\"game_id\":\"$GAME_ID\",\"row\":1,\"col\":0}" \
    $SERVER_ADDR tictactoe.TicTacToeService/MakeMove

echo "9. Player 1 makes a move..."
grpcurl -plaintext -d "{\"user_id\":\"player1\",\"game_id\":\"$GAME_ID\",\"row\":2,\"col\":2}" \
    $SERVER_ADDR tictactoe.TicTacToeService/MakeMove

echo "10. Getting game state... player 1"
grpcurl -plaintext -d "{\"user_id\":\"player1\",\"game_id\":\"$GAME_ID\"}" \
    $SERVER_ADDR tictactoe.TicTacToeService/GetGame

echo "10. Getting game state... player 2"
grpcurl -plaintext -d "{\"user_id\":\"player2\",\"game_id\":\"$GAME_ID\"}" \
    $SERVER_ADDR tictactoe.TicTacToeService/GetGame

echo "11. Getting user stats... player 1"
grpcurl -plaintext -d '{"user_id":"player1"}' \
    $SERVER_ADDR tictactoe.TicTacToeService/GetUserStats

echo "11. Getting user stats... player 2"
grpcurl -plaintext -d '{"user_id":"player2"}' \
    $SERVER_ADDR tictactoe.TicTacToeService/GetUserStats

echo "✓ All tests completed successfully!"
