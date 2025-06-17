package core

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		allowedOrigin := os.Getenv("ALLOWED_WS_ORIGIN")
		
		// Allow production domain
		if strings.HasPrefix(origin, allowedOrigin) {
			return true
		}
		
		// Allow Docker internal networks, and development ports
		if strings.HasPrefix(origin, "http://172.") { // Docker internal network
			return true
		}
		
		log.Printf("Rejected WebSocket connection from origin: %s", origin)
		return false
	},
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OllamaRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
	Options  struct {
		NumCtx int `json:"num_ctx"`
	} `json:"options"`
}

type OllamaResponse struct {
	Model     string `json:"model"`
	CreatedAt string `json:"created_at"`
	Message   struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
	Done           bool   `json:"done"`
	Response       string `json:"response"` // For streaming
	DoneReason     string `json:"done_reason"`
	TotalDuration  int64  `json:"total_duration"`
}

func HandleChat(w http.ResponseWriter, r *http.Request, conn *pgx.Conn, user User) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer ws.Close()

	// Get user's CV as context
	cvContext := ""
	if user.CV != "" {
		cvContext = "Here is the user's CV for context:\n" + user.CV + "\n\n"
	}

	// Original system prompt without think tag instructions
	systemPrompt := cvContext + `You are an expert team builder assistant. Help the user form effective teams based on their project requirements. 
	Ask clarifying questions if needed and provide insightful suggestions.`

	conversation := []ChatMessage{
		{Role: "system", Content: systemPrompt},
	}

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			break
		}

		// Add user message to conversation
		userMsg := string(message)
		conversation = append(conversation, ChatMessage{Role: "user", Content: userMsg})

		// Call Ollama API with streaming
		ollamaReq := OllamaRequest{
			Model:    os.Getenv("OLLAMA_MODEL"),
			Messages: conversation,
			Stream:   true, 
		}

		numCTX, err := strconv.Atoi(os.Getenv("OLLAMA_CTX"))
		if err != nil {
			log.Println("Could not convert the OLLAMA_CTX env variable to int.")
			numCTX = 4096
		}
		ollamaReq.Options.NumCtx = numCTX // Set context length

		// Make streaming request to Ollama
		reqBody, _ := json.Marshal(ollamaReq)
		ollamaAPI := os.Getenv("OLLAMA_API")
		if ollamaAPI == "" {
			ollamaAPI = "http://localhost:11434/api/chat"
		}
		resp, err := http.Post(
			ollamaAPI,
			"application/json", 
			bytes.NewReader(reqBody),
		)
		if err != nil {
			log.Printf("Ollama API error: %v", err)
			ws.WriteMessage(websocket.TextMessage, []byte("Sorry, I'm having trouble connecting to the assistant."))
			continue
		}
		defer resp.Body.Close()

		// Create a buffer to accumulate the assistant's response
		var assistantResponseBuilder strings.Builder

		// Stream the response line by line
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Bytes()
			var chunk OllamaResponse
			if err := json.Unmarshal(line, &chunk); err != nil {
				log.Printf("Error parsing Ollama response: %v", err)
				continue
			}

			// Skip empty responses
			if chunk.Response == "" && chunk.Message.Content == "" {
				continue
			}

			// Get content from either field
			content := chunk.Response
			if content == "" {
				content = chunk.Message.Content
			}

			// Add to buffer
			assistantResponseBuilder.WriteString(content)

			// Send chunk to client
			if err := ws.WriteMessage(websocket.TextMessage, []byte(content)); err != nil {
				log.Printf("WebSocket write error: %v", err)
				break
			}

			// Break if this is the last chunk
			if chunk.Done {
				break
			}
		}

		if err := scanner.Err(); err != nil {
			log.Printf("Error reading Ollama response: %v", err)
		}

		// Add full assistant response to conversation
		fullResponse := assistantResponseBuilder.String()
		conversation = append(conversation, ChatMessage{
			Role:    "assistant",
			Content: fullResponse,
		})
	}
}
