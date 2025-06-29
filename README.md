# Gemini 2.0 API Chat Server

A Go-based API server that provides a chat interface using Google's Gemini 2.0 API with conversation context management and session handling.

## Features

- **Session Management**: Maintains conversation context with a 60-second inactivity timeout
- **Base Prompt Integration**: Automatically prepends a base prompt to all conversations
- **Gemini 2.0 API Integration**: Uses Google's Gemini 2.0 Flash Experimental model
- **Context Preservation**: Keeps conversation history for contextual responses
- **Automatic Cleanup**: Removes expired sessions to manage memory usage
- **Discord Integration**: Sends logs and conversation events to Discord webhook

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

## Usage

### Starting the Server
```bash
# Set your Gemini API key
export GEMINI_API_KEY="your_api_key_here"

# Basic usage (defaults to localhost:8080)
go run main.go config.go

# Custom host and port
go run main.go config.go -h 192.168.1.100 -p 9000

# With custom base prompt file
go run main.go config.go -prompt /path/to/prompt.txt

# All options together
go run main.go config.go -h 0.0.0.0 -p 8080 -prompt custom_prompt.txt
```

### Command Line Options
- `-h`: Host to bind the server to (default: "localhost")
- `-p`: Port to bind the server to (default: "8080")
- `-prompt`: Path to file containing custom base prompt (optional, uses config.go default if not provided)

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

### Testing
Run the test client to see the API in action:
```bash
go run client_test.go
```

## Configuration

- **Host**: Configurable via `-h` flag (default: "localhost")
- **Port**: Configurable via `-p` flag (default: "8080")
- **Base Prompt**: Configurable via `-prompt` flag or defined in `config.go` as fallback
- **Session Timeout**: 60 seconds of inactivity
- **Model**: gemini-2.0-flash-exp (Google Gemini 2.0)
- **Context Limit**: Maximum 20 messages per session
- **API Key**: Loaded from `GEMINI_API_KEY` environment variable
- **Logging**: Timestamped logs with request/response tracking
- **Discord Webhook**: Integrated for real-time logging and conversation monitoring

## Session Management

- Sessions are created automatically when a new `session_id` is provided
- Each session maintains conversation history
- Sessions expire after 60 seconds of inactivity
- Expired sessions are automatically cleaned up every 30 seconds
- Context is limited to the last 20 messages to manage API costs

## Security Notes

- The API key is now loaded from the `GEMINI_API_KEY` environment variable
- Set the environment variable before running the server: `export GEMINI_API_KEY="your_api_key_here"`
- To get a Gemini API key, visit: https://makersuite.google.com/app/apikey
- Consider adding authentication and rate limiting
- Validate and sanitize user inputs

## Dependencies

- Standard Go libraries only (no external dependencies)
- Requires Go 1.16 or later
