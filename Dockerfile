# Use official Golang image as base
FROM golang:1.22-alpine

# Set working directory
WORKDIR /app

# Copy Go source files to /app in container
COPY main.go config.go start.sh ./

# Make startup script executable
RUN chmod +x start.sh

# Set environment variable for API key (can be overridden with --env on run)
ENV GEMINI_API_KEY="AIzaSyCGrwGPzWY3W90ZFHgfGGdX5Azj3g7rFAE"

# Build the Go binary
RUN go build -o chatbot main.go config.go

# Expose port (Railway will override this)
EXPOSE 8008

# Use startup script for better debugging
CMD ["./start.sh"]
