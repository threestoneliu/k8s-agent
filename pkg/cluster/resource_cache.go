package cluster

import (
	"fmt"
	"strings"
	"sync"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
)

// ResourceCache provides cached access to API resource metadata per cluster.
// It lazily fetches resources from DiscoveryClient on first access and caches them.
type ResourceCache struct {
	mu      sync.RWMutex
	caches  map[string]*clusterResourceCache
	ttl     time.Duration
}

// clusterResourceCache holds cached resource info for a single cluster
type clusterResourceCache struct {
	gvrIndex         map[string]schema.GroupVersionResource // resource name -> GVR
	gvrIndexByShort  map[string]schema.GroupVersionResource // short name -> GVR
	namespacedIndex  map[string]bool                         // resource name -> is namespaced
	apiResources     []metav1.APIResource
	lastRefresh      time.Time
}

// NewResourceCache creates a new ResourceCache with the given TTL for cached entries
func NewResourceCache(ttl time.Duration) *ResourceCache {
	return &ResourceCache{
		caches: make(map[string]*clusterResourceCache),
		ttl:    ttl,
	}
}

// GetGVR returns the GroupVersionResource for a given resource name (supports short names).
// Returns false if the resource is not found.
func (rc *ResourceCache) GetGVR(clusterName, resource string) (schema.GroupVersionResource, bool) {
	cache := rc.getClusterCache(clusterName)
	if cache == nil {
		return schema.GroupVersionResource{}, false
	}

	// Try exact match first
	gvr, ok := cache.gvrIndex[resource]
	if ok {
		return gvr, true
	}

	// Try short name match
	gvr, ok = cache.gvrIndexByShort[resource]
	if ok {
		return gvr, true
	}

	return schema.GroupVersionResource{}, false
}

// IsNamespaced returns true if the resource is namespaced, false if cluster-scoped.
// Returns false if the resource is not found.
func (rc *ResourceCache) IsNamespaced(clusterName, resource string) bool {
	cache := rc.getClusterCache(clusterName)
	if cache == nil {
		return false
	}

	namespaced, ok := cache.namespacedIndex[resource]
	if !ok {
		// Try short name
		for short, gvr := range cache.gvrIndexByShort {
			if short == resource {
				namespaced, ok = cache.namespacedIndex[gvr.Resource]
				return ok && namespaced
			}
		}
		return false
	}
	return namespaced
}

// Refresh forces a refresh of the resource cache for a cluster
func (rc *ResourceCache) Refresh(clusterName string, client kubernetes.Interface) error {
	return rc.refresh(clusterName, client)
}

// GetAPIResources returns the cached API resources for a cluster
func (rc *ResourceCache) GetAPIResources(clusterName string) []metav1.APIResource {
	cache := rc.getClusterCache(clusterName)
	if cache == nil {
		return nil
	}
	return cache.apiResources
}

func (rc *ResourceCache) getClusterCache(clusterName string) *clusterResourceCache {
	rc.mu.RLock()
	defer rc.mu.RUnlock()

	cache, ok := rc.caches[clusterName]
	if !ok {
		return nil
	}

	// Check if cache is still valid
	if rc.ttl > 0 && time.Since(cache.lastRefresh) > rc.ttl {
		return nil
	}

	return cache
}

func (rc *ResourceCache) refresh(clusterName string, client kubernetes.Interface) error {
	if client == nil {
		return fmt.Errorf("kubernetes client is nil")
	}

	discoveryClient := client.Discovery()
	_, apiResourceLists, err := discoveryClient.ServerGroupsAndResources()
	if err != nil {
		return fmt.Errorf("failed to get API resources: %w", err)
	}

	cache := &clusterResourceCache{
		gvrIndex:        make(map[string]schema.GroupVersionResource),
		gvrIndexByShort: make(map[string]schema.GroupVersionResource),
		namespacedIndex: make(map[string]bool),
		apiResources:    make([]metav1.APIResource, 0),
	}

	for _, apiList := range apiResourceLists {
		for _, resource := range apiList.APIResources {
			// Use apiList.GroupVersion for Version when resource.Version is empty
			version := resource.Version
			if version == "" {
				// Split "group/version" to get just the version part (e.g., "rbac.authorization.k8s.io/v1" -> "v1")
				if parts := strings.Split(apiList.GroupVersion, "/"); len(parts) == 2 {
					version = parts[1]
				} else {
					version = apiList.GroupVersion
				}
			}
			// Extract Group from apiList.GroupVersion (e.g., "rbac.authorization.k8s.io/v1" -> "rbac.authorization.k8s.io")
			group := ""
			if apiList.GroupVersion != "v1" {
				if parts := strings.Split(apiList.GroupVersion, "/"); len(parts) == 2 {
					group = parts[0]
				}
			}
			gvr := schema.GroupVersionResource{
				Group:    group,
				Version:  version,
				Resource: resource.Name,
			}

			// Index by primary name
			cache.gvrIndex[resource.Name] = gvr
			cache.namespacedIndex[resource.Name] = resource.Namespaced

			// Index short names
			for _, short := range resource.ShortNames {
				cache.gvrIndexByShort[short] = gvr
			}
		}
	}

	// Collect all API resources for GetAPIResources
	for _, apiList := range apiResourceLists {
		cache.apiResources = append(cache.apiResources, apiList.APIResources...)
	}

	cache.lastRefresh = time.Now()

	// Note: caller holds the lock, no need to re-acquire
	rc.caches[clusterName] = cache

	return nil
}

// ensureCache ensures the cache is populated for a cluster, refreshing if needed
func (rc *ResourceCache) ensureCache(clusterName string, client kubernetes.Interface) error {
	rc.mu.RLock()
	_, exists := rc.caches[clusterName]
	rc.mu.RUnlock()

	if !exists {
		return rc.refresh(clusterName, client)
	}

	return nil
}

// LazyRefresh populates the cache for a cluster if not already populated
func (rc *ResourceCache) LazyRefresh(clusterName string, client kubernetes.Interface) error {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	// Check if cache exists and is still valid under write lock
	if rc.caches[clusterName] != nil {
		c := rc.caches[clusterName]
		if rc.ttl <= 0 || time.Since(c.lastRefresh) <= rc.ttl {
			return nil // Cache is valid, no refresh needed
		}
	}

	return rc.refresh(clusterName, client)
}
