package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s-agent/pkg/engine"
)

func TestConfirmCommandExecution(t *testing.T) {
	cmd := NewRootCommand()

	// Test with confirm key that doesn't exist
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"confirm", "non-existent-key"})

	err := cmd.Execute()
	assert.Error(t, err) // Should error because key doesn't exist
}

func TestConfirmCommand_MissingArgs(t *testing.T) {
	cmd := NewRootCommand()
	confirmCmd, _, err := cmd.Find([]string{"confirm"})
	require.NoError(t, err)

	// Test with missing args
	buf := &bytes.Buffer{}
	confirmCmd.SetOut(buf)
	confirmCmd.SetErr(buf)

	err = confirmCmd.RunE(confirmCmd, []string{}) // Missing confirm key
	assert.Error(t, err)
}

// MockConfirmManager is a mock for confirmation operations
type MockConfirmManager struct {
	confirmations map[string]*MockConfirmOperation
}

type MockConfirmOperation struct {
	ConfirmKey    string
	TargetCluster string
	Operation     *engine.ClassifiedOperation
	Status        string
}

func NewMockConfirmManager() *MockConfirmManager {
	return &MockConfirmManager{
		confirmations: make(map[string]*MockConfirmOperation),
	}
}

func (m *MockConfirmManager) Create(key string, cluster string, op *engine.ClassifiedOperation) string {
	confirmKey := "mock-key-" + key
	m.confirmations[confirmKey] = &MockConfirmOperation{
		ConfirmKey:    confirmKey,
		TargetCluster: cluster,
		Operation:     op,
		Status:        "pending",
	}
	return confirmKey
}

func (m *MockConfirmManager) Approve(key string) error {
	if op, exists := m.confirmations[key]; exists {
		op.Status = "approved"
		return nil
	}
	return nil
}

func (m *MockConfirmManager) Get(key string) (*MockConfirmOperation, error) {
	if op, exists := m.confirmations[key]; exists {
		return op, nil
	}
	return nil, nil
}

func TestMockConfirmManager_Create(t *testing.T) {
	mock := NewMockConfirmManager()

	op := &engine.ClassifiedOperation{
		Type:     engine.OperationTypeMutation,
		Verb:     "delete",
		Resource: "pods",
		Name:     "my-pod",
	}

	key := mock.Create("test-1", "test-cluster", op)
	assert.NotEmpty(t, key)

	// Verify created
	retrieved, err := mock.Get(key)
	require.NoError(t, err)
	assert.Equal(t, "test-cluster", retrieved.TargetCluster)
	assert.Equal(t, "pending", retrieved.Status)
}

func TestMockConfirmManager_Approve(t *testing.T) {
	mock := NewMockConfirmManager()

	op := &engine.ClassifiedOperation{
		Type:     engine.OperationTypeMutation,
		Verb:     "delete",
		Resource: "pods",
		Name:     "my-pod",
	}

	key := mock.Create("test-1", "test-cluster", op)

	// Approve
	err := mock.Approve(key)
	require.NoError(t, err)

	// Verify approved
	retrieved, err := mock.Get(key)
	require.NoError(t, err)
	assert.Equal(t, "approved", retrieved.Status)
}

func TestMockConfirmManager_GetNonExistent(t *testing.T) {
	mock := NewMockConfirmManager()

	retrieved, err := mock.Get("non-existent")
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestConfirmKeyValidation(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		isValid bool
	}{
		{
			name:    "valid key",
			key:     "abc123def456",
			isValid: true,
		},
		{
			name:    "short key",
			key:     "abc",
			isValid: false,
		},
		{
			name:    "empty key",
			key:     "",
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := validateConfirmKey(tt.key)
			assert.Equal(t, tt.isValid, isValid)
		})
	}
}

// Helper to validate confirm key
func validateConfirmKey(key string) bool {
	if key == "" {
		return false
	}
	if len(key) < 8 {
		return false
	}
	return true
}

func TestConfirmSubCommandExecution(t *testing.T) {
	cmd := NewRootCommand()

	// Test confirm without args
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetArgs([]string{"confirm"})

	err := cmd.Execute()
	assert.Error(t, err)
}
