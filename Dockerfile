# Use official Golang image as base
FROM golang:1.22 as builder

# Set working directory
WORKDIR /app

# Copy Go source files to /app in container
COPY main.go config.go ./

# Make startup script executable

# Set environment variable for API key (can be overridden with --env on run)
ENV GEMINI_API_KEY="AIzaSyCGrwGPzWY3W90ZFHgfGGdX5Azj3g7rFAE"

# Build the Go binary
RUN go build -o chatbot main.go config.go

# Expose port (Railway will override this)
EXPOSE 8008

# Use startup script for better debugging
CMD ["./chatbot","-h","0.0.0.0","-p","8008","-t","120"]
