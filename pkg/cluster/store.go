package cluster

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// Config paths
var (
	DefaultClustersPath = filepath.Join(os.Getenv("HOME"), ".config", "k8s-agent", "clusters.yaml")
	DefaultConfigPath   = filepath.Join(os.Getenv("HOME"), ".config", "k8s-agent", "config.yaml")
)

// Store provides persistent storage for cluster configurations
type Store struct {
	configPath   string
	clustersPath string
	mu           sync.RWMutex
}

// StoreOption is a functional option for Store
type StoreOption func(*Store)

// WithConfigPath sets the config file path
func WithConfigPath(path string) StoreOption {
	return func(s *Store) {
		s.configPath = path
	}
}

// WithClustersPath sets the clusters file path
func WithClustersPath(path string) StoreOption {
	return func(s *Store) {
		s.clustersPath = path
	}
}

// appConfig represents the config.yaml structure
type appConfig struct {
	CurrentCluster string `yaml:"current-cluster"`
}

// clustersConfig represents the clusters.yaml structure
type clustersConfig struct {
	Clusters []*ClusterConfig `yaml:"clusters"`
}

// NewStore creates a new Store instance
func NewStore(opts ...StoreOption) (*Store, error) {
	s := &Store{
		configPath:   DefaultConfigPath,
		clustersPath: DefaultClustersPath,
	}

	for _, opt := range opts {
		opt(s)
	}

	// Ensure directory exists
	if err := s.ensureDir(); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	return s, nil
}

// ensureDir ensures the config directory exists
func (s *Store) ensureDir() error {
	// Ensure directory for config file
	if err := os.MkdirAll(filepath.Dir(s.configPath), 0755); err != nil {
		return err
	}
	// Ensure directory for clusters file
	return os.MkdirAll(filepath.Dir(s.clustersPath), 0755)
}

// SaveCluster saves a cluster configuration
func (s *Store) SaveCluster(cfg *ClusterConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	clusters, err := s.loadClustersLocked()
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to load clusters: %w", err)
	}

	// Update or add cluster
	found := false
	for i, c := range clusters {
		if c.Name == cfg.Name {
			clusters[i] = cfg
			found = true
			break
		}
	}
	if !found {
		clusters = append(clusters, cfg)
	}

	// Save clusters
	if err := s.saveClustersLocked(clusters); err != nil {
		return fmt.Errorf("failed to save clusters: %w", err)
	}

	return nil
}

// LoadCluster loads a cluster configuration by name
func (s *Store) LoadCluster(name string) (*ClusterConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clusters, err := s.loadClustersLocked()
	if err != nil {
		return nil, err
	}

	for _, c := range clusters {
		if c.Name == name {
			return c, nil
		}
	}

	return nil, os.ErrNotExist
}

// ListClusters returns all cluster configurations
func (s *Store) ListClusters() []*ClusterConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clusters, err := s.loadClustersLocked()
	if err != nil {
		return []*ClusterConfig{}
	}

	// Return a copy to prevent external mutation
	result := make([]*ClusterConfig, len(clusters))
	copy(result, clusters)
	return result
}

// DeleteCluster deletes a cluster configuration
func (s *Store) DeleteCluster(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	clusters, err := s.loadClustersLocked()
	if err != nil {
		if os.IsNotExist(err) {
			return nil // Idempotent delete
		}
		return err
	}

	// Filter out the cluster to delete
	found := false
	newClusters := make([]*ClusterConfig, 0, len(clusters))
	for _, c := range clusters {
		if c.Name == name {
			found = true
			continue
		}
		newClusters = append(newClusters, c)
	}

	if !found {
		// Cluster didn't exist, but that's okay - idempotent delete
		return nil
	}

	return s.saveClustersLocked(newClusters)
}

// SetCurrentCluster sets the current cluster context
func (s *Store) SetCurrentCluster(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	config := &appConfig{CurrentCluster: name}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(s.configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config: %w", err)
	}

	return nil
}

// GetCurrentCluster returns the current cluster context
func (s *Store) GetCurrentCluster() (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	var config appConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return "", fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return config.CurrentCluster, nil
}

// loadClustersLocked loads clusters from file (caller must hold lock)
func (s *Store) loadClustersLocked() ([]*ClusterConfig, error) {
	data, err := os.ReadFile(s.clustersPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []*ClusterConfig{}, nil
		}
		return nil, err
	}

	var config clustersConfig
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal clusters: %w", err)
	}

	if config.Clusters == nil {
		return []*ClusterConfig{}, nil
	}

	return config.Clusters, nil
}

// saveClustersLocked saves clusters to file (caller must hold lock)
func (s *Store) saveClustersLocked(clusters []*ClusterConfig) error {
	config := clustersConfig{Clusters: clusters}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal clusters: %w", err)
	}

	if err := os.WriteFile(s.clustersPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write clusters: %w", err)
	}

	return nil
}
