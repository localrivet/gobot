package tools

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// GrepTool searches for patterns in files
type GrepTool struct{}

// NewGrepTool creates a new grep tool
func NewGrepTool() *GrepTool {
	return &GrepTool{}
}

// Name returns the tool name
func (t *GrepTool) Name() string {
	return "grep"
}

// Description returns the tool description
func (t *GrepTool) Description() string {
	return `Search for patterns in files using regular expressions.
Returns matching lines with file paths and line numbers.
Use glob parameter to filter which files to search.`
}

// Schema returns the JSON schema for the tool input
func (t *GrepTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"pattern": {
				"type": "string",
				"description": "Regular expression pattern to search for"
			},
			"path": {
				"type": "string",
				"description": "File or directory to search in (default: current directory)"
			},
			"glob": {
				"type": "string",
				"description": "Glob pattern to filter files (e.g., '*.go', '**/*.ts')"
			},
			"case_insensitive": {
				"type": "boolean",
				"description": "Make search case-insensitive (default: false)"
			},
			"context": {
				"type": "integer",
				"description": "Number of lines of context around matches (default: 0)"
			},
			"limit": {
				"type": "integer",
				"description": "Maximum number of matches to return (default: 100)"
			}
		},
		"required": ["pattern"]
	}`)
}

// GrepInput represents the tool input
type GrepInput struct {
	Pattern         string `json:"pattern"`
	Path            string `json:"path"`
	Glob            string `json:"glob"`
	CaseInsensitive bool   `json:"case_insensitive"`
	Context         int    `json:"context"`
	Limit           int    `json:"limit"`
}

// GrepMatch represents a single match
type GrepMatch struct {
	File    string
	Line    int
	Content string
}

// Execute searches for the pattern
func (t *GrepTool) Execute(ctx context.Context, input json.RawMessage) (*ToolResult, error) {
	var in GrepInput
	if err := json.Unmarshal(input, &in); err != nil {
		return nil, fmt.Errorf("invalid input: %w", err)
	}

	if in.Pattern == "" {
		return &ToolResult{
			Content: "Error: pattern is required",
			IsError: true,
		}, nil
	}

	// Set defaults
	if in.Path == "" {
		in.Path = "."
	}
	if in.Limit <= 0 {
		in.Limit = 100
	}

	// Expand home directory
	if strings.HasPrefix(in.Path, "~/") {
		home, _ := os.UserHomeDir()
		in.Path = filepath.Join(home, in.Path[2:])
	}

	// Compile regex
	flags := ""
	if in.CaseInsensitive {
		flags = "(?i)"
	}
	re, err := regexp.Compile(flags + in.Pattern)
	if err != nil {
		return &ToolResult{
			Content: fmt.Sprintf("Invalid regex pattern: %v", err),
			IsError: true,
		}, nil
	}

	// Get files to search
	var files []string
	info, err := os.Stat(in.Path)
	if err != nil {
		return &ToolResult{
			Content: fmt.Sprintf("Error: %v", err),
			IsError: true,
		}, nil
	}

	if info.IsDir() {
		files, err = t.findFiles(in.Path, in.Glob)
		if err != nil {
			return &ToolResult{
				Content: fmt.Sprintf("Error finding files: %v", err),
				IsError: true,
			}, nil
		}
	} else {
		files = []string{in.Path}
	}

	// Search files
	var matches []GrepMatch
	matchCount := 0

	for _, file := range files {
		if matchCount >= in.Limit {
			break
		}

		fileMatches, err := t.searchFile(file, re, in.Context, in.Limit-matchCount)
		if err != nil {
			continue // Skip files we can't read
		}

		matches = append(matches, fileMatches...)
		matchCount += len(fileMatches)
	}

	if len(matches) == 0 {
		return &ToolResult{
			Content: fmt.Sprintf("No matches found for pattern: %s", in.Pattern),
		}, nil
	}

	// Format output
	var result strings.Builder
	for _, m := range matches {
		result.WriteString(fmt.Sprintf("%s:%d: %s\n", m.File, m.Line, m.Content))
	}

	if matchCount >= in.Limit {
		result.WriteString(fmt.Sprintf("\n... (showing first %d matches)", in.Limit))
	}

	return &ToolResult{
		Content: strings.TrimSpace(result.String()),
	}, nil
}

// findFiles finds all files matching the glob in the directory
func (t *GrepTool) findFiles(dir, glob string) ([]string, error) {
	var files []string

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() {
			// Skip hidden and common non-source directories
			name := info.Name()
			if strings.HasPrefix(name, ".") && name != "." {
				return filepath.SkipDir
			}
			if name == "node_modules" || name == "vendor" || name == "__pycache__" {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip binary files by extension
		ext := filepath.Ext(path)
		binaryExts := map[string]bool{
			".exe": true, ".bin": true, ".so": true, ".dylib": true,
			".png": true, ".jpg": true, ".gif": true, ".ico": true,
			".zip": true, ".tar": true, ".gz": true,
		}
		if binaryExts[ext] {
			return nil
		}

		// Check glob pattern if specified
		if glob != "" {
			matched, _ := filepath.Match(glob, info.Name())
			if !matched {
				return nil
			}
		}

		files = append(files, path)

		// Limit files to search
		if len(files) >= 10000 {
			return filepath.SkipAll
		}
		return nil
	})

	return files, err
}

// searchFile searches a single file for the pattern
func (t *GrepTool) searchFile(path string, re *regexp.Regexp, contextLines, maxMatches int) ([]GrepMatch, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var matches []GrepMatch
	var lines []string
	scanner := bufio.NewScanner(file)

	// Read all lines if we need context
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	for lineNum, line := range lines {
		if len(matches) >= maxMatches {
			break
		}

		if re.MatchString(line) {
			// Truncate long lines
			content := line
			if len(content) > 500 {
				content = content[:500] + "..."
			}

			matches = append(matches, GrepMatch{
				File:    path,
				Line:    lineNum + 1,
				Content: content,
			})
		}
	}

	return matches, nil
}

// RequiresApproval returns false - searching is safe
func (t *GrepTool) RequiresApproval() bool {
	return false
}
