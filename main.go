package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

const groqAPIURL = "https://api.groq.com/openai/v1/chat/completions"

// Configuration
type Config struct {
	Host           string
	Port           string
	BasePrompt     string
	LogFile        string
	SessionTimeout time.Duration
}

var config Config
var fileLogger *log.Logger
var logFile *os.File

// Load base prompt from file or use default
func loadBasePrompt(promptFile string) string {
	if promptFile == "" {
		return BasePrompt // from config.go
	}

	data, err := os.ReadFile(promptFile)
	if err != nil {
		fmt.Printf("Warning: Could not read prompt file '%s': %v\n", promptFile, err)
		fmt.Printf("Using default base prompt from config.go\n")
		return BasePrompt
	}

	prompt := string(data)
	if prompt == "" {
		fmt.Printf("Warning: Prompt file '%s' is empty, using default\n", promptFile)
		return BasePrompt
	}

	fmt.Printf("Loaded base prompt from file: %s\n", promptFile)
	return prompt
}

// Initialize logging
func initLogging(logFilePath string) error {
	if logFilePath == "" {
		logFilePath = "chatbot.log"
	}

	var err error
	logFile, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	fileLogger = log.New(logFile, "", log.LstdFlags)
	config.LogFile = logFilePath

	fmt.Printf("Logging to file: %s\n", logFilePath)
	return nil
}

// Logger with timestamp - logs to both stdout and file
func logf(format string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf("[%s] %s", timestamp, fmt.Sprintf(format, args...))

	// Log to stdout
	fmt.Println(message)

	// Log to file
	if fileLogger != nil {
		fileLogger.Println(message)
	}
}

// Close logging
func closeLogging() {
	if logFile != nil {
		logFile.Close()
	}
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Choice struct {
	Message Message `json:"message"`
}

type ChatResponse struct {
	Choices []Choice `json:"choices"`
}

// Session represents a conversation session
type Session struct {
	ID       string    `json:"id"`
	History  []Message `json:"history"`
	LastSeen time.Time `json:"last_seen"`
	Mutex    sync.RWMutex
}

// API request/response structures
type PromptRequest struct {
	SessionID string `json:"session_id"`
	Prompt    string `json:"prompt"`
}

type PromptResponse struct {
	SessionID string `json:"session_id"`
	Response  string `json:"response"`
	Error     string `json:"error,omitempty"`
}

// Global session manager
type SessionManager struct {
	sessions map[string]*Session
	mutex    sync.RWMutex
	ticker   *time.Ticker
	done     chan bool
}

var sessionManager *SessionManager

func NewSessionManager() *SessionManager {
	sm := &SessionManager{
		sessions: make(map[string]*Session),
		ticker:   time.NewTicker(30 * time.Second), // Check every 30 seconds
		done:     make(chan bool),
	}

	// Start cleanup goroutine
	go sm.cleanupRoutine()

	return sm
}

func (sm *SessionManager) cleanupRoutine() {
	for {
		select {
		case <-sm.ticker.C:
			sm.cleanupExpiredSessions()
		case <-sm.done:
			sm.ticker.Stop()
			return
		}
	}
}

func (sm *SessionManager) cleanupExpiredSessions() {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	now := time.Now()
	cleanedCount := 0
	for id, session := range sm.sessions {
		session.Mutex.RLock()
		lastSeen := session.LastSeen
		session.Mutex.RUnlock()

		if now.Sub(lastSeen) > config.SessionTimeout {
			delete(sm.sessions, id)
			cleanedCount++
		}
	}
	if cleanedCount > 0 {
		logf("Cleaned up %d expired session(s) (timeout: %v)", cleanedCount, config.SessionTimeout)
	}
}

func (sm *SessionManager) getOrCreateSession(sessionID string) *Session {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	session, exists := sm.sessions[sessionID]
	if !exists {
		session = &Session{
			ID:       sessionID,
			History:  []Message{{Role: "system", Content: config.BasePrompt}},
			LastSeen: time.Now(),
		}
		sm.sessions[sessionID] = session
		logf("Created new session: %s", sessionID)
	} else {
		session.Mutex.Lock()
		session.LastSeen = time.Now()
		session.Mutex.Unlock()
		logf("Session accessed: %s", sessionID)
	}

	return session
}

func callGroqAPI(apiKey string, messages []Message) (string, error) {
	reqBody := ChatRequest{
		Model:    "llama3-8b-8192", // Using the currently available free tier model
		Messages: messages,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", groqAPIURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBytes))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBytes, &chatResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %v", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	return chatResp.Choices[0].Message.Content, nil
}

func handlePrompt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		logf("Invalid method %s for /prompt endpoint", r.Method)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PromptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logf("Invalid JSON in request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.SessionID == "" {
		logf("Missing session ID in request")
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	if req.Prompt == "" {
		logf("Missing prompt in request from session %s", req.SessionID)
		http.Error(w, "Prompt is required", http.StatusBadRequest)
		return
	}

	logf("Processing prompt from session %s: %s", req.SessionID, truncateString(req.Prompt, 50))

	// Get or create session
	session := sessionManager.getOrCreateSession(req.SessionID)

	session.Mutex.Lock()
	defer session.Mutex.Unlock()

	// Add user message to history
	session.History = append(session.History, Message{Role: "user", Content: req.Prompt})

	// Call Groq API
	apiKey := "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"
	response, err := callGroqAPI(apiKey, session.History)

	var resp PromptResponse
	resp.SessionID = req.SessionID

	if err != nil {
		logf("API error for session %s: %v", req.SessionID, err)
		resp.Error = err.Error()
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		logf("Generated response for session %s: %s", req.SessionID, truncateString(response, 50))
		resp.Response = response
		// Add assistant response to history
		session.History = append(session.History, Message{Role: "assistant", Content: response})

		// Keep only last 20 messages to manage context length
		if len(session.History) > 20 {
			// Keep system message and last 19 messages
			session.History = append(session.History[:1], session.History[len(session.History)-19:]...)
			logf("Truncated history for session %s to 20 messages", req.SessionID)
		}
	}

	// Log full conversation details to file
	logConversation(req.SessionID, req.Prompt, response, err)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	logf("Health check request from %s", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
}

// Helper function to truncate strings for logging
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// Log full conversation details to file
func logConversation(sessionID, userPrompt, aiResponse string, err error) {
	if fileLogger == nil {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")

	fileLogger.Printf("=== CONVERSATION LOG [%s] ===", timestamp)
	fileLogger.Printf("Session ID: %s", sessionID)
	fileLogger.Printf("User Prompt: %s", userPrompt)

	if err != nil {
		fileLogger.Printf("Error: %v", err)
	} else {
		fileLogger.Printf("AI Response: %s", aiResponse)
	}
	fileLogger.Printf("=== END CONVERSATION ===\n")
}

func main() {
	// Parse command line flags
	var host = flag.String("h", "localhost", "Host to bind the server to")
	var port = flag.String("p", "8080", "Port to bind the server to")
	var promptFile = flag.String("prompt", "", "Path to file containing base prompt (optional)")
	var logFile = flag.String("log", "chatbot.log", "Path to log file")
	var timeout = flag.Int("t", 60, "Session timeout in seconds")
	flag.Parse()

	// Initialize logging
	if err := initLogging(*logFile); err != nil {
		fmt.Printf("Failed to initialize logging: %v\n", err)
		os.Exit(1)
	}
	defer closeLogging()

	// Load base prompt
	config.BasePrompt = loadBasePrompt(*promptFile)
	config.Host = *host
	config.Port = *port
	config.SessionTimeout = time.Duration(*timeout) * time.Second

	// Initialize session manager
	sessionManager = NewSessionManager()
	defer func() {
		sessionManager.done <- true
	}()

	// Set up routes
	http.HandleFunc("/prompt", handlePrompt)
	http.HandleFunc("/health", handleHealth)

	// Start server
	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)
	logf("Starting Groq API Chat Server")
	logf("Server address: %s", addr)
	logf("Base prompt loaded: %d characters", len(config.BasePrompt))
	logf("Session timeout: %v", config.SessionTimeout)
	logf("Model: llama3-8b-8192 (Groq free tier)")
	logf("Log file: %s", config.LogFile)
	logf("Press Ctrl+C to stop the server")

	if err := http.ListenAndServe(addr, nil); err != nil {
		logf("Server error: %v", err)
		os.Exit(1)
	}
}
