# Use official Golang image as base
FROM golang:1.22 as builder

# Set working directory
WORKDIR /app

# Copy Go source files to /app in container
COPY main.go config.go ./

# Make startup script executable

# Set environment variable for API key (can be overridden with --env on run)
ENV GEMINI_API_KEY="AIzaSyDykWow1uQNwQf6Hvdq_tMgX46IbplLTjk"
ENV DISCORD_WEBHOOK_URL="https://discord.com/api/webhooks/1390603709263777812/hrgrbP95DFRvgSkqCtjhMGMar75hs_zW-02fbptMcSWlRPLEsL9tlsBekwtvtHCZds-n"

# Build the Go binary
RUN go build -o chatbot main.go config.go

# Expose port (Railway will override this)
EXPOSE 8008

# Use startup script for better debugging
CMD ["./chatbot","-h","0.0.0.0","-p","8080","-t","120"]
