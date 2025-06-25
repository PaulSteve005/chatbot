# Groq API Chat Server

A Go-based API server that provides a chat interface using the Groq API with conversation context management and session handling.

## Features

- **Session Management**: Maintains conversation context with a 60-second inactivity timeout
- **Base Prompt Integration**: Automatically prepends a base prompt to all conversations
- **Groq API Integration**: Uses Groq's free tier API (llama3-8b-8192 model)
- **Context Preservation**: Keeps conversation history for contextual responses
- **Automatic Cleanup**: Removes expired sessions to manage memory usage

## API Endpoints

### POST /prompt
Send a prompt and receive a response with conversation context.

**Request Body:**
```json
{
  "session_id": "unique-session-id",
  "prompt": "Your message here"
}
```

**Response:**
```json
{
  "session_id": "unique-session-id",
  "response": "AI response here",
  "error": "error message if any"
}
```

### GET /health
Health check endpoint.

**Response:**
```json
{
  "status": "healthy"
}
```

### GET /stats
Server statistics endpoint.

**Response:**
```json
{
  "active_sessions": 5,
  "total_sessions": 25,
  "expired_sessions": 20,
  "session_timeout": "2m0s",
  "uptime": "1h23m45s"
}
```

## Usage

### Starting the Server
```bash
# Basic usage (defaults to localhost:8080, 60s timeout)
go run main.go

# Custom host and port
go run main.go -h 192.168.1.100 -p 9000

# With custom session timeout (30 seconds)
go run main.go -t 30

# With custom base prompt file
go run main.go -prompt /path/to/prompt.txt

# All options together
go run main.go -h 0.0.0.0 -p 8080 -prompt custom_prompt.txt -t 120 -log custom.log
```

### Command Line Options
- `-h`: Host to bind the server to (default: "localhost")
- `-p`: Port to bind the server to (default: "8080")
- `-prompt`: Path to file containing custom base prompt (optional, uses config.go default if not provided)
- `-log`: Path to log file (default: "chatbot.log")
- `-t`: Session timeout in seconds (default: 60)

The server will start on the specified host and port.

### Example API Calls

Using curl:
```bash
# First message in a session
curl -X POST http://localhost:8080/prompt \
  -H "Content-Type: application/json" \
  -d '{"session_id": "user123", "prompt": "Hello! What can you help me with?"}'

# Follow-up message (maintains context)
curl -X POST http://localhost:8080/prompt \
  -H "Content-Type: application/json" \
  -d '{"session_id": "user123", "prompt": "Can you remember what I just asked?"}'
```


## Configuration

- **Host**: Configurable via `-h` flag (default: "localhost")
- **Port**: Configurable via `-p` flag (default: "8080")
- **Base Prompt**: Configurable via `-prompt` flag or defined in `config.go` as fallback
- **Log File**: Configurable via `-log` flag (default: "chatbot.log")
- **Session Timeout**: Configurable via `-t` flag (default: 60 seconds)
- **Model**: llama3-8b-8192 (Groq free tier)
- **Context Limit**: Maximum 20 messages per session
- **API Key**: Currently hardcoded (should be moved to environment variable in production)
- **Logging**: Timestamped logs with request/response tracking and full conversation details

## Session Management

- Sessions are created automatically when a new `session_id` is provided
- Each session maintains conversation history
- Sessions expire after 60 seconds of inactivity
- Expired sessions are automatically cleaned up every 30 seconds
- Context is limited to the last 20 messages to manage API costs

## Security Notes

- The API key is currently hardcoded in the source code
- For production use, move the API key to an environment variable
- Consider adding authentication and rate limiting
- Validate and sanitize user inputs

## Dependencies

- Standard Go libraries only (no external dependencies)
- Requires Go 1.16 or later
