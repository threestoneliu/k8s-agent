package engine

import "strings"

// resourceMappings maps various resource name forms to canonical API resource names
var resourceMappings = map[string]string{
	// Pods
	"pod":        "pods",
	"pods":       "pods",
	// Deployments
	"deployment": "deployments",
	"deployments": "deployments",
	"deploy":     "deployments",
	// Services
	"service":    "services",
	"services":   "services",
	"svc":        "services",
	"svcs":       "services",
	// Namespaces
	"namespace":  "namespaces",
	"namespaces": "namespaces",
	"ns":         "namespaces",
	// Nodes
	"node":       "nodes",
	"nodes":      "nodes",
	// ConfigMaps
	"configmap":  "configmaps",
	"configmaps": "configmaps",
	"cm":         "configmaps",
	// Secrets
	"secret":     "secrets",
	"secrets":    "secrets",
	// Ingress
	"ingress":    "ingresses",
	"ingresses":  "ingresses",
	"ing":        "ingresses",
	// PersistentVolumes
	"persistentvolume":       "persistentvolumes",
	"persistentvolumes":     "persistentvolumes",
	"pv":                    "persistentvolumes",
	"persistentvolumeclaim": "persistentvolumeclaims",
	"persistentvolumeclaims": "persistentvolumeclaims",
	"pvc":                   "persistentvolumeclaims",
	// StorageClasses
	"storageclass":      "storageclasses",
	"storageclasses":   "storageclasses",
	"sc":               "storageclasses",
	// Endpoints
	"endpoint":  "endpoints",
	"endpoints": "endpoints",
	"ep":        "endpoints",
	// Events
	"event":  "events",
	"events": "events",
	// ResourceQuotas
	"resourcequota":     "resourcequotas",
	"resourcequotas":    "resourcequotas",
	"quota":            "resourcequotas",
	// LimitRanges
	"limitrange":    "limitranges",
	"limitranges":   "limitranges",
	"limits":        "limitranges",
	// HorizontalPodAutoscaler
	"horizontalpodautoscaler": "horizontalpodautoscalers",
	"horizontalpodautoscalers": "horizontalpodautoscalers",
	"hpa":                    "horizontalpodautoscalers",
	// PodDisruptionBudget
	"poddisruptionbudget": "poddisruptionbudgets",
	"poddisruptionbudgets": "poddisruptionbudgets",
	"pdb":                  "poddisruptionbudgets",
}

// MapResource maps a resource name (including aliases) to its canonical Kubernetes API resource name
func MapResource(resource string) string {
	if mapped, ok := resourceMappings[strings.ToLower(resource)]; ok {
		return mapped
	}
	return resource
}

// IsValidResource checks if a resource name is a known Kubernetes resource type
func IsValidResource(resource string) bool {
	canonical := MapResource(resource)
	_, ok := resourceMappings[strings.ToLower(canonical)]
	return ok
}
