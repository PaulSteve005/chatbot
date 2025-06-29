#!/bin/bash

# Set your Gemini API key here
export GEMINI_API_KEY="your_api_key_here"

# Build the server
go build -o chatbot main.go config.go

# Start the server in background
./chatbot -h 0.0.0.0 -p 9000 -t 120 &
SERVER_PID=$!

# Wait for server to start
sleep 3

# Test the API
echo "Testing Gemini 2.0 API server..."
curl -X POST http://0.0.0.0:9000/prompt \
  -H "Content-Type: application/json" \
  -d '{"session_id": "test123", "prompt": "Hello! What can you help me with?"}'

echo -e "\n\nTesting health endpoint..."
curl http://0.0.0.0:9000/health

# Stop the server
kill $SERVER_PID


