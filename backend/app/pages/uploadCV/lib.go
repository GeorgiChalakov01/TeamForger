package uploadCV

import (
	"teamforger/backend/core"
	"github.com/jackc/pgx/v5"
	"context"

	"regexp"
	"sort"
	"strings"
	"fmt"
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

func StoreUserCV(conn *pgx.Conn, user core.User) error {
	fmt.Println(user.CV)
	chunks := chunkCV(user.CV)
	for i := 0; i < len(chunks); i++ {
		fmt.Println("==================================")
		fmt.Println(chunks[i])
	}
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

	return nil
}

