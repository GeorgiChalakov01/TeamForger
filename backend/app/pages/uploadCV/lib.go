package uploadCV

import (
	"teamforger/backend/core"
	"github.com/jackc/pgx/v5"
	"context"

	"regexp"
	"sort"
	"strings"
	"fmt"

	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/pgvector/pgvector-go"
)

func chunkCV(cv string) []string {
	sectionHeaders := []string{
		"PROFESSIONAL SUMMARY",
		"TECHNICAL EXPERTISE/KNOWLEDGE",
		"PROJECT HIGHLIGHTS",
		"TRAININGS",
		"EDUCATION",
		"CERTIFICATIONS",
		"LANGUAGES",
	}

	upperCV := strings.ToUpper(cv)
	var sectionsFound []struct {
		header string
		start  int
	}
	for _, header := range sectionHeaders {
		pattern := `(?m)^[^a-zA-Z0-9]*` + regexp.QuoteMeta(header) + `[^a-zA-Z0-9]*$`
		re := regexp.MustCompile(pattern)
		loc := re.FindStringIndex(upperCV)
		if loc != nil {
			sectionsFound = append(sectionsFound, struct {
				header string
				start  int
			}{header: header, start: loc[0]})
		}
	}

	// Handle case with no sections found
	if len(sectionsFound) == 0 {
		return []string{strings.TrimSpace(cv)}
	}

	// Sort sections by position
	sort.Slice(sectionsFound, func(i, j int) bool {
		return sectionsFound[i].start < sectionsFound[j].start
	})

	var chunks []string

	// Add initial chunk before first section
	if sectionsFound[0].start > 0 {
		initialChunk := strings.TrimSpace(cv[:sectionsFound[0].start])
		if initialChunk != "" {
			chunks = append(chunks, initialChunk)
		}
	}

	// Process each section
	for i, sec := range sectionsFound {
		start := sec.start
		end := len(cv)
		if i < len(sectionsFound)-1 {
			end = sectionsFound[i+1].start
		}
		chunk := cv[start:end]

		// Remove header line
		if idx := strings.Index(chunk, "\n"); idx != -1 {
			chunk = strings.TrimSpace(chunk[idx+1:])
		} else {
			chunk = ""
		}

		if sec.header == "PROJECT HIGHLIGHTS" {
			projects := splitProjects(chunk)
			for _, p := range projects {
				trimmed := strings.TrimSpace(p)
				if trimmed != "" {
					// Add prefix to each project
					chunks = append(chunks, "Worked in project: "+trimmed)
				}
			}
		} else {
			if chunk != "" {
				chunks = append(chunks, chunk)
			}
		}
	}

	return chunks
}

// Split project highlights into individual projects
func splitProjects(chunk string) []string {
	if chunk == "" {
		return nil
	}

	// Use a state machine to properly identify projects
	var projects []string
	var currentProject strings.Builder
	inProject := false
	lines := strings.Split(chunk, "\n")

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		upperLine := strings.ToUpper(trimmed)
		
		// Detect project start (line with parentheses/dates indicating a project title)
		if isProjectStart(upperLine, i, lines) {
			if inProject && currentProject.Len() > 0 {
				// Save current project before starting new one
				projects = append(projects, currentProject.String())
				currentProject.Reset()
			}
			inProject = true
		}

		if inProject {
			currentProject.WriteString(line)
			currentProject.WriteString("\n")
		}

		// Detect environment section end
		if strings.HasPrefix(upperLine, "ENVIRONMENT:") && isEnvironmentEnd(i, lines) {
			if inProject && currentProject.Len() > 0 {
				projects = append(projects, currentProject.String())
				currentProject.Reset()
				inProject = false
			}
		}
	}

	// Add last project if exists
	if currentProject.Len() > 0 {
		projects = append(projects, currentProject.String())
	}

	return projects
}

// Heuristic to detect project start lines
func isProjectStart(upperLine string, currentIndex int, lines []string) bool {
	// Typical project titles contain parentheses and dates
	hasParentheses := strings.Contains(upperLine, "(") && strings.Contains(upperLine, ")")
	hasDates := strings.Contains(upperLine, "–") || strings.Contains(upperLine, "-")
	
	// Check if previous line might be a project header
	if currentIndex > 0 {
		prevLine := strings.TrimSpace(lines[currentIndex-1])
		if strings.Contains(prevLine, "Projects:") {
			return true
		}
	}
	
	return hasParentheses && hasDates
}

// Check if environment section has ended
func isEnvironmentEnd(currentIndex int, lines []string) bool {
	if currentIndex >= len(lines)-1 {
		return true
	}
	
	// Environment section ends when we hit an empty line or a new project
	nextLine := strings.TrimSpace(lines[currentIndex+1])
	if nextLine == "" {
		return true
	}
	
	// Next line looks like a project start
	if currentIndex+1 < len(lines) && isProjectStart(strings.ToUpper(nextLine), currentIndex+1, lines) {
		return true
	}
	
	return false
}

type EmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}

func GetEmbedding(text string) ([]float32, error) {
	// Get configuration from environment variables
	model := os.Getenv("OLLAMA_EMB_MODEL")
	if model == "" {
		model = "nomic-embed-text"
		return nil, fmt.Errorf("OLLAMA_EMB_MODEL environment variable not set")
	}

	apiURL := os.Getenv("OLLAMA_EMB_API")
	if apiURL == "" {
		apiURL = "http://localhost:11434/api/embeddings" // Default if not set
	}

	// Create request payload
	payload := map[string]string{
		"model":  model,
		"prompt": text,
	}

	// Marshal payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("error marshaling JSON: %w", err)
	}

	// Create HTTP request with timeout
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request with timeout
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	// Check for successful response
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error [%d]: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var embeddingResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embeddingResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return embeddingResp.Embedding, nil
}

func StoreUserCV(conn *pgx.Conn, user core.User) error {
	// Store the original CV.
	// Start a transaction
	tx, err := conn.Begin(context.Background())
	if err != nil {
		return err
	}
	// Rollback is safe to call even if the tx is already closed, so if
	// the tx commits successfully, this is a no-op
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), "UPDATE users SET cv = $1 WHERE email = $2", user.CV, user.Email)

	if err != nil {
		return err
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return err
	}

	// Insert the chunked CV.
	chunks := chunkCV(user.CV)
	for i := 0; i < len(chunks); i++ {
		embeddingFA, err := GetEmbedding(chunks[i])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return err
		}
		embedding := pgvector.NewVector(embeddingFA)

		// Start a transaction
		tx, err := conn.Begin(context.Background())
		if err != nil {
			return err
		}
		// Rollback is safe to call even if the tx is already closed, so if
		// the tx commits successfully, this is a no-op
		defer tx.Rollback(context.Background())

		_, err = tx.Exec(context.Background(), "INSERT INTO cv_chunks (user_id, chunk, embedding) VALUES ($1, $2, $3)", user.Id, chunks[i], embedding)

		if err != nil {
			return err
		}

		err = tx.Commit(context.Background())
		if err != nil {
			return err
		}
	}
	return nil
}

