// Package skills provides a YAML-based skill definition and loading system.
// Skills are declarative definitions that modify agent behavior without requiring
// compiled code - just YAML/Markdown files that can be hot-reloaded.
package skills

import (
	"fmt"
	"strings"
)

// Skill represents a declarative skill definition that modifies agent behavior.
// Skills can add context to prompts, require specific tools, and provide examples.
type Skill struct {
	// Name is the unique identifier for the skill
	Name string `yaml:"name"`

	// Description explains what the skill does
	Description string `yaml:"description"`

	// Version for tracking skill updates
	Version string `yaml:"version"`

	// Triggers are keywords/phrases that activate this skill
	Triggers []string `yaml:"triggers"`

	// Template is additional system prompt content when skill is active
	Template string `yaml:"template"`

	// Tools lists required tool names for this skill
	Tools []string `yaml:"tools"`

	// Examples provide few-shot learning examples
	Examples []Example `yaml:"examples"`

	// Priority determines precedence when multiple skills match (higher = first)
	Priority int `yaml:"priority"`

	// Enabled allows disabling skills without removing them
	Enabled bool `yaml:"enabled"`

	// FilePath stores where this skill was loaded from
	FilePath string `yaml:"-"`
}

// Example represents a user-assistant exchange for few-shot learning
type Example struct {
	User      string `yaml:"user"`
	Assistant string `yaml:"assistant"`
}

// Matches checks if the given input text triggers this skill
func (s *Skill) Matches(input string) bool {
	if !s.Enabled {
		return false
	}

	inputLower := strings.ToLower(input)
	for _, trigger := range s.Triggers {
		if strings.Contains(inputLower, strings.ToLower(trigger)) {
			return true
		}
	}
	return false
}

// ApplyToPrompt modifies the system prompt with skill-specific content
func (s *Skill) ApplyToPrompt(systemPrompt string) string {
	if s.Template == "" && len(s.Examples) == 0 {
		return systemPrompt
	}

	var sb strings.Builder
	sb.WriteString(systemPrompt)

	// Add skill template
	if s.Template != "" {
		sb.WriteString("\n\n## Skill: ")
		sb.WriteString(s.Name)
		sb.WriteString("\n")
		sb.WriteString(s.Template)
	}

	// Add examples as few-shot prompts
	if len(s.Examples) > 0 {
		sb.WriteString("\n\n### Examples for ")
		sb.WriteString(s.Name)
		sb.WriteString(":\n")
		for i, ex := range s.Examples {
			sb.WriteString(fmt.Sprintf("\n**Example %d:**\n", i+1))
			sb.WriteString("User: ")
			sb.WriteString(ex.User)
			sb.WriteString("\nAssistant: ")
			sb.WriteString(ex.Assistant)
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// RequiredTools returns the list of tools this skill needs
func (s *Skill) RequiredTools() []string {
	return s.Tools
}

// Validate checks if the skill definition is valid
func (s *Skill) Validate() error {
	if s.Name == "" {
		return fmt.Errorf("skill name is required")
	}
	if s.Description == "" {
		return fmt.Errorf("skill %q: description is required", s.Name)
	}
	return nil
}
