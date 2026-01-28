package skills

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSkillMatches(t *testing.T) {
	skill := &Skill{
		Name:        "test-skill",
		Description: "A test skill",
		Triggers:    []string{"review", "check code"},
		Enabled:     true,
	}

	tests := []struct {
		input   string
		matches bool
	}{
		{"please review my code", true},
		{"can you check code for bugs", true},
		{"REVIEW this file", true}, // Case insensitive
		{"hello world", false},
		{"run tests", false},
	}

	for _, tt := range tests {
		result := skill.Matches(tt.input)
		if result != tt.matches {
			t.Errorf("Matches(%q) = %v, want %v", tt.input, result, tt.matches)
		}
	}
}

func TestSkillMatchesDisabled(t *testing.T) {
	skill := &Skill{
		Name:     "disabled-skill",
		Triggers: []string{"hello"},
		Enabled:  false,
	}

	if skill.Matches("hello world") {
		t.Error("Disabled skill should not match")
	}
}

func TestSkillApplyToPrompt(t *testing.T) {
	skill := &Skill{
		Name:        "test-skill",
		Description: "Test",
		Template:    "When reviewing code, look for bugs.",
		Examples: []Example{
			{User: "Review this", Assistant: "I'll check for issues."},
		},
		Enabled: true,
	}

	result := skill.ApplyToPrompt("Base prompt")

	if result == "Base prompt" {
		t.Error("ApplyToPrompt should modify the prompt")
	}

	if len(result) <= len("Base prompt") {
		t.Error("ApplyToPrompt should add content to prompt")
	}
}

func TestSkillValidate(t *testing.T) {
	tests := []struct {
		skill   Skill
		wantErr bool
	}{
		{Skill{Name: "test", Description: "Test"}, false},
		{Skill{Name: "", Description: "Test"}, true},      // Missing name
		{Skill{Name: "test", Description: ""}, true},      // Missing description
		{Skill{}, true},                                    // Empty
	}

	for _, tt := range tests {
		err := tt.skill.Validate()
		if (err != nil) != tt.wantErr {
			t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
		}
	}
}

func TestLoaderLoadAll(t *testing.T) {
	// Create temp directory with test skills
	dir := t.TempDir()

	skillYAML := `name: test-skill
description: A test skill for testing
version: "1.0.0"
triggers:
  - test
  - testing
template: |
  This is a test template.
`
	err := os.WriteFile(filepath.Join(dir, "test.yaml"), []byte(skillYAML), 0644)
	if err != nil {
		t.Fatal(err)
	}

	loader := NewLoader(dir)
	if err := loader.LoadAll(); err != nil {
		t.Fatalf("LoadAll() error = %v", err)
	}

	if loader.Count() != 1 {
		t.Errorf("Count() = %d, want 1", loader.Count())
	}

	skill, ok := loader.Get("test-skill")
	if !ok {
		t.Fatal("Get() failed to find skill")
	}

	if skill.Name != "test-skill" {
		t.Errorf("skill.Name = %q, want %q", skill.Name, "test-skill")
	}
}

func TestLoaderFindMatching(t *testing.T) {
	dir := t.TempDir()

	skill1YAML := `name: skill-1
description: First skill
triggers:
  - hello
priority: 10
`
	skill2YAML := `name: skill-2
description: Second skill
triggers:
  - world
priority: 5
`
	skill3YAML := `name: skill-3
description: Third skill
triggers:
  - hello
priority: 20
`

	os.WriteFile(filepath.Join(dir, "skill1.yaml"), []byte(skill1YAML), 0644)
	os.WriteFile(filepath.Join(dir, "skill2.yaml"), []byte(skill2YAML), 0644)
	os.WriteFile(filepath.Join(dir, "skill3.yaml"), []byte(skill3YAML), 0644)

	loader := NewLoader(dir)
	loader.LoadAll()

	// Test matching "hello" - should return 2 skills, highest priority first
	matching := loader.FindMatching("hello there")
	if len(matching) != 2 {
		t.Errorf("FindMatching() returned %d skills, want 2", len(matching))
	}

	if matching[0].Name != "skill-3" {
		t.Errorf("First match should be skill-3 (priority 20), got %s", matching[0].Name)
	}

	// Test matching "world" - should return 1 skill
	matching = loader.FindMatching("hello world")
	if len(matching) != 3 {
		t.Errorf("FindMatching('hello world') returned %d skills, want 3", len(matching))
	}
}

func TestLoaderEmptyDir(t *testing.T) {
	loader := NewLoader("/nonexistent/path")
	err := loader.LoadAll()
	if err != nil {
		t.Errorf("LoadAll() should not error for nonexistent dir, got %v", err)
	}
	if loader.Count() != 0 {
		t.Errorf("Count() = %d, want 0 for empty/nonexistent dir", loader.Count())
	}
}
