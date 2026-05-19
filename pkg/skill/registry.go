package skill

import (
	"fmt"
	"sync"
)

// Registry manages all loaded Skills
type Registry struct {
	mu     sync.RWMutex
	skills map[string]*Skill
}

// NewRegistry creates a new empty Registry
func NewRegistry() *Registry {
	return &Registry{
		skills: make(map[string]*Skill),
	}
}

// Register adds a Skill to the registry
func (r *Registry) Register(s *Skill) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.skills[s.Name]; exists {
		return fmt.Errorf("skill %q already registered", s.Name)
	}
	r.skills[s.Name] = s
	return nil
}

// Get retrieves a Skill by name
func (r *Registry) Get(name string) (*Skill, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	s, ok := r.skills[name]
	return s, ok
}

// List returns all registered Skills
func (r *Registry) List() []*Skill {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*Skill, 0, len(r.skills))
	for _, s := range r.skills {
		result = append(result, s)
	}
	return result
}

// Count returns the number of registered Skills
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.skills)
}