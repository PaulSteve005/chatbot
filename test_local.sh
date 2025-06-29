#!/bin/bash

echo "=== Local Test Script ==="

# Set test API key (replace with your actual key)
export GEMINI_API_KEY="your_test_api_key_here"

# Build the application
echo "Building application..."
go build -o chatbot main.go config.go

if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

echo "Build successful!"

# Start server in background
echo "Starting server..."
./chatbot -h 0.0.0.0 -p 8008 -t 120 &
SERVER_PID=$!

# Wait for server to start
sleep 3

# Test endpoints
echo "Testing root endpoint..."
curl -s http://localhost:8008/ | jq .

echo "Testing health endpoint..."
curl -s http://localhost:8008/health | jq .

echo "Testing prompt endpoint..."
curl -s -X POST http://localhost:8008/prompt \
  -H "Content-Type: application/json" \
  -d '{"session_id": "test123", "prompt": "Hello! What can you help me with?"}' | jq .

# Stop server
kill $SERVER_PID
echo "Test completed!" 