package engine

import (
	"errors"

	"k8s-agent/pkg/log"
)

// executeMutation returns confirmation required for mutation operations
func (e *Executor) executeMutation(clusterName string, op *ClassifiedOperation) (*ExecutionResult, error) {
	log.Info("mutation operation requires confirmation", "cluster", clusterName, "verb", op.Verb, "resource", op.Resource, "name", op.Name)
	return &ExecutionResult{
		Success:   false,
		Output:    "confirmation_required",
		Error:     errors.New("confirmation required for mutation operation"),
		Resource:  op.Resource,
		Verb:      op.Verb,
	}, nil
}
