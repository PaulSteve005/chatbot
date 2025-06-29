#!/bin/bash

echo "=== Railway Startup Script ==="
echo "Current directory: $(pwd)"
echo "Files in directory:"
ls -la

echo ""
echo "Environment variables:"
echo "PORT: $PORT"
echo "GEMINI_API_KEY: ${GEMINI_API_KEY:0:10}..."

echo ""
echo "Building application..."
go build -o chatbot main.go config.go

if [ $? -eq 0 ]; then
    echo "Build successful!"
    echo "Starting server on 0.0.0.0:${PORT:-8008}..."
    ./chatbot -h 0.0.0.0 -p ${PORT:-8008} -t 120
else
    echo "Build failed!"
    exit 1
fi 