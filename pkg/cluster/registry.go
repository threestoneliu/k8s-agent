package cluster

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// Error variables
var (
	ErrClusterNotFound      = errors.New("cluster not found")
	ErrClusterAlreadyExists = errors.New("cluster already exists")
	ErrInvalidClusterName   = errors.New("invalid cluster name")
	ErrInvalidKubeconfig    = errors.New("invalid kubeconfig")
)

// ClusterConfig holds the configuration for a managed cluster
type ClusterConfig struct {
	Name       string
	Kubeconfig string
}

// Registry manages multiple Kubernetes cluster connections
type Registry struct {
	mu       sync.RWMutex
	store    *Store
	clusters map[string]*ClusterConfig
	clients  map[string]kubernetes.Interface
}

// RegistryOption is a functional option for Registry
type RegistryOption func(*Registry)

// WithStore sets the store for persistence
func WithStore(store *Store) RegistryOption {
	return func(r *Registry) {
		r.store = store
	}
}

// NewRegistry creates a new cluster registry
func NewRegistry(opts ...RegistryOption) *Registry {
	r := &Registry{
		clusters: make(map[string]*ClusterConfig),
		clients:  make(map[string]kubernetes.Interface),
	}

	for _, opt := range opts {
		opt(r)
	}

	// If store is set, load clusters from it
	if r.store != nil {
		r.loadFromStore()
	}

	return r
}

// loadFromStore loads clusters from the store into memory
func (r *Registry) loadFromStore() {
	clusters := r.store.ListClusters()
	for _, cfg := range clusters {
		r.clusters[cfg.Name] = cfg
	}
}

// GetCluster returns the kubernetes client for a given cluster name
func (r *Registry) GetCluster(name string) (kubernetes.Interface, error) {
	return r.GetClusterContext(context.Background(), name)
}

// GetClusterContext returns the kubernetes client for a given cluster name with context support
func (r *Registry) GetClusterContext(ctx context.Context, name string) (kubernetes.Interface, error) {
	if name == "" {
		return nil, ErrClusterNotFound
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	cfg, ok := r.clusters[name]
	if !ok {
		return nil, ErrClusterNotFound
	}

	client, ok := r.clients[name]
	if !ok {
		// Lazy load the client
		client, err := r.buildClient(cfg.Kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build client: %w", err)
		}
		r.clients[name] = client
		return client, nil
	}

	return client, nil
}

// ListClusterNames returns all configured cluster names
func (r *Registry) ListClusterNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]string, 0, len(r.clusters))
	for name := range r.clusters {
		result = append(result, name)
	}
	return result
}

// ListClusters returns all configured clusters
func (r *Registry) ListClusters() []*ClusterConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*ClusterConfig, 0, len(r.clusters))
	for _, cfg := range r.clusters {
		result = append(result, cfg)
	}
	return result
}

// AddCluster adds a new cluster to the registry
func (r *Registry) AddCluster(name, kubeconfig string) error {
	if name == "" {
		return ErrInvalidClusterName
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.clusters[name]; exists {
		return ErrClusterAlreadyExists
	}

	cfg := &ClusterConfig{
		Name:       name,
		Kubeconfig: kubeconfig,
	}

	r.clusters[name] = cfg

	// Persist to store if available
	if r.store != nil {
		if err := r.store.SaveCluster(cfg); err != nil {
			return fmt.Errorf("failed to persist cluster: %w", err)
		}
	}

	return nil
}

// RemoveCluster removes a cluster from the registry
func (r *Registry) RemoveCluster(name string) error {
	if name == "" {
		return ErrClusterNotFound
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.clusters[name]; !exists {
		return ErrClusterNotFound
	}

	delete(r.clusters, name)
	delete(r.clients, name)

	// Persist to store if available
	if r.store != nil {
		if err := r.store.DeleteCluster(name); err != nil {
			return fmt.Errorf("failed to persist cluster deletion: %w", err)
		}
	}

	return nil
}

// GetCurrentCluster returns the name of the current cluster
func (r *Registry) GetCurrentCluster() (string, error) {
	if r.store == nil {
		return "", nil
	}
	return r.store.GetCurrentCluster()
}

// SetCurrentCluster sets the current cluster context
func (r *Registry) SetCurrentCluster(name string) error {
	if r.store == nil {
		return fmt.Errorf("no store configured")
	}
	return r.store.SetCurrentCluster(name)
}

// buildClient builds a kubernetes client from a kubeconfig path
func (r *Registry) buildClient(kubeconfigPath string) (kubernetes.Interface, error) {
	if kubeconfigPath == "" {
		// Use default kubeconfig
		kubeconfigPath = os.Getenv("KUBECONFIG")
		if kubeconfigPath == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return nil, fmt.Errorf("failed to get home directory: %w", err)
			}
			kubeconfigPath = homeDir + "/.kube/config"
		}
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidKubeconfig, err)
	}

	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %w", err)
	}

	return client, nil
}
