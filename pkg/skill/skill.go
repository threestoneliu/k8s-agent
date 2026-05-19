package skill

// Skill represents a loaded Skill with metadata and location
type Skill struct {
	Name        string
	Description string
	Location    string // absolute path to SKILL.md
}

// Frontmatter represents the YAML frontmatter of a SKILL.md
type Frontmatter struct {
	Name           string            `yaml:"name"`
	Description    string            `yaml:"description"`
	License        string            `yaml:"license,omitempty"`
	Compatibility  string            `yaml:"compatibility,omitempty"`
	Metadata       map[string]string `yaml:"metadata,omitempty"`
}

// String returns the skill name
func (s *Skill) String() string {
	return s.Name
}