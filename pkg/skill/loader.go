package skill

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// LoadSkills loads all Skills from the user config directory
func LoadSkills(configDir string) ([]*Skill, error) {
	skillsDir := filepath.Join(configDir, "skills")

	// Check if directory exists
	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		return []*Skill{}, nil // empty is ok
	}

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read skills directory: %w", err)
	}

	var skills []*Skill
	for _, entry := range entries {
		if !entry.IsDir() {
			continue // skip files
		}

		skillPath := filepath.Join(skillsDir, entry.Name(), "SKILL.md")
		skill, err := loadSkillFile(skillPath)
		if err != nil {
			// log warning but continue
			fmt.Printf("Warning: failed to load skill %s: %v\n", entry.Name(), err)
			continue
		}
		skills = append(skills, skill)
	}

	return skills, nil
}

// loadSkillFile loads a single SKILL.md file
func loadSkillFile(path string) (*Skill, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read file: %w", err)
	}

	// Split frontmatter from markdown body
	fm, _, err := parseFrontmatter(string(data))
	if err != nil {
		return nil, fmt.Errorf("invalid frontmatter: %w", err)
	}

	// Validate required fields
	if fm.Name == "" {
		return nil, fmt.Errorf("missing required field: name")
	}
	if fm.Description == "" {
		return nil, fmt.Errorf("missing required field: description")
	}

	// Resolve absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve absolute path: %w", err)
	}

	return &Skill{
		Name:        fm.Name,
		Description: fm.Description,
		Location:    absPath,
	}, nil
}

// parseFrontmatter extracts YAML frontmatter from markdown content
func parseFrontmatter(content string) (*Frontmatter, string, error) {
	if len(content) < 4 || content[:4] != "---\n" {
		return nil, "", fmt.Errorf("missing frontmatter separator")
	}

	endIdx := 4
	for i := 4; i < len(content)-3; i++ {
		if content[i:i+4] == "\n---" {
			endIdx = i + 4
			break
		}
	}

	yamlContent := content[4 : endIdx-4]
	body := content[endIdx:]

	var fm Frontmatter
	if err := yaml.Unmarshal([]byte(yamlContent), &fm); err != nil {
		return nil, "", fmt.Errorf("invalid yaml: %w", err)
	}

	return &fm, body, nil
}