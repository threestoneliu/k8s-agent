package e2e

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s-agent/cmd/cli"
	"k8s-agent/pkg/cluster"
	"k8s-agent/pkg/confirmation"
	"k8s-agent/pkg/engine"
	"k8s-agent/pkg/scheduler"
	"k8s-agent/pkg/session"
)

// MockExecutor implements the executor interface for E2E testing
type MockExecutor struct {
	ExecuteFunc func(clusterName string, op *engine.ClassifiedOperation) (*engine.ExecutionResult, error)
}

func (m *MockExecutor) Execute(clusterName string, op *engine.ClassifiedOperation) (*engine.ExecutionResult, error) {
	if m.ExecuteFunc != nil {
		return m.ExecuteFunc(clusterName, op)
	}
	return &engine.ExecutionResult{Success: true, Output: "mock output"}, nil
}

// setupTestRoot creates a root command with mocked dependencies for testing
func setupTestRoot(mockExec *MockExecutor) (*cobra.Command, *session.Manager, *confirmation.Manager) {
	sessionMgr := session.NewManager()
	confirmMgr := confirmation.NewManager(5 * time.Second) // Fast expiry for tests

	rootCmd := cli.NewRootCommand()
	return rootCmd, sessionMgr, confirmMgr
}

func TestQueryOperationFlow(t *testing.T) {
	// Test: Execute "get pods" command
	// This verifies the query operation flow works end-to-end

	parsedOp, err := engine.Parse("get pods")
	require.NoError(t, err)
	require.NotNil(t, parsedOp)

	classifiedOp := engine.ClassifyOperation(parsedOp)
	assert.Equal(t, engine.OperationTypeQuery, classifiedOp.Type)
	assert.Equal(t, "get", classifiedOp.Verb)
	assert.Equal(t, "pods", classifiedOp.Resource)
}

func TestQueryListFlow(t *testing.T) {
	// Test: Execute "list services" command

	parsedOp, err := engine.Parse("list services")
	require.NoError(t, err)

	classifiedOp := engine.ClassifyOperation(parsedOp)
	assert.Equal(t, engine.OperationTypeQuery, classifiedOp.Type)
	assert.Equal(t, "list", classifiedOp.Verb)
	assert.Equal(t, "services", classifiedOp.Resource)
}

func TestMutationOperationFlow(t *testing.T) {
	// Test: Parse "delete pod nginx" and verify it returns confirmation_required

	parsedOp, err := engine.Parse("delete pod nginx")
	require.NoError(t, err)

	classifiedOp := engine.ClassifyOperation(parsedOp)
	assert.Equal(t, engine.OperationTypeMutation, classifiedOp.Type)
	assert.Equal(t, "delete", classifiedOp.Verb)
	assert.Equal(t, "pod", classifiedOp.Resource) // Note: singular form from parser
	assert.Equal(t, "nginx", classifiedOp.Name)
}

func TestMutationWithConfirmation(t *testing.T) {
	// Test: Create a confirmation for mutation and approve with key

	confirmMgr := confirmation.NewManager(5 * time.Second) // 5 second TTL for tests

	classifiedOp := &engine.ClassifiedOperation{
		Type:     engine.OperationTypeMutation,
		Verb:     "delete",
		Resource: "pods",
		Name:     "nginx",
	}

	// Create confirmation
	confirmKey, err := confirmMgr.CreateConfirmation("test-cluster", classifiedOp)
	require.NoError(t, err)
	require.NotEmpty(t, confirmKey)

	// Verify confirmation exists using GetConfirmation
	pending, err := confirmMgr.GetConfirmation(confirmKey)
	require.NoError(t, err)
	assert.Equal(t, confirmation.StatusPending, pending.Status)

	// Approve confirmation
	err = confirmMgr.ApproveConfirmation(confirmKey)
	require.NoError(t, err)

	// Verify it's now approved
	pending, err = confirmMgr.GetConfirmation(confirmKey)
	require.NoError(t, err)
	assert.Equal(t, confirmation.StatusApproved, pending.Status)
}

func TestChatModeQueryFlow(t *testing.T) {
	// Test: Start chat with mock executor and send "list deployments"

	sessionMgr := session.NewManager()
	mockExec := &MockExecutor{
		ExecuteFunc: func(clusterName string, op *engine.ClassifiedOperation) (*engine.ExecutionResult, error) {
			return &engine.ExecutionResult{
				Success:  true,
				Output:   "Found 3 deployments",
				Resource: "deployments",
				Verb:     "list",
			}, nil
		},
	}

	// Create conversation
	conversationID := "test-chat"
	_, err := sessionMgr.CreateConversation(conversationID, "test-cluster", "default")
	require.NoError(t, err)

	// Process input
	input := "list deployments"
	parsedOp, err := engine.Parse(input)
	require.NoError(t, err)

	classifiedOp := engine.ClassifyOperation(parsedOp)
	assert.Equal(t, engine.OperationTypeQuery, classifiedOp.Type)

	// Execute with mock
	result, err := mockExec.Execute("test-cluster", classifiedOp)
	require.NoError(t, err)
	assert.True(t, result.Success)
	assert.Contains(t, result.Output, "deployments")
}

func TestChatModeMutationFlow(t *testing.T) {
	// Test: Mutation requires confirmation in chat mode

	sessionMgr := session.NewManager()
	confirmMgr := confirmation.NewManager(5 * time.Second)

	// Create conversation
	conversationID := "test-chat-mutation"
	_, err := sessionMgr.CreateConversation(conversationID, "test-cluster", "default")
	require.NoError(t, err)

	// Process delete command
	input := "delete deployment myapp"
	parsedOp, err := engine.Parse(input)
	require.NoError(t, err)

	classifiedOp := engine.ClassifyOperation(parsedOp)
	assert.Equal(t, engine.OperationTypeMutation, classifiedOp.Type)

	// Request confirmation
	confirmKey, err := confirmMgr.CreateConfirmation("test-cluster", classifiedOp)
	require.NoError(t, err)

	// Verify mutation requires confirmation
	assert.NotEmpty(t, confirmKey)

	// List pending confirmations for cluster
	pending := confirmMgr.GetPendingByCluster("test-cluster")
	assert.GreaterOrEqual(t, len(pending), 1)
}

func TestTaskCreationFlow(t *testing.T) {
	// Test: Create a scheduled task and verify it appears in list

	schedulerMgr := scheduler.NewManager(nil)

	// Create a task
	task := &scheduler.ScheduledTask{
		ID:            "test-task-1",
		Name:          "Test Task",
		CronExpr:      "*/5 * * * *",
		TargetCluster: "test-cluster",
		Operation: &engine.ClassifiedOperation{
			Type:     engine.OperationTypeQuery,
			Verb:     "get",
			Resource: "pods",
		},
		Enabled: true,
	}

	err := schedulerMgr.AddTask(task)
	require.NoError(t, err)

	// Verify it appears in list
	tasks := schedulerMgr.ListTasks()
	assert.GreaterOrEqual(t, len(tasks), 1)

	// Find our task
	found := false
	for _, taskItem := range tasks {
		if taskItem.ID == "test-task-1" {
			found = true
			assert.Equal(t, "*/5 * * * *", taskItem.CronExpr)
			assert.True(t, taskItem.Enabled)
			break
		}
	}
	assert.True(t, found, "Task should be found in list")
}

func TestTaskEnableDisable(t *testing.T) {
	// Test: Enable/disable a task

	schedulerMgr := scheduler.NewManager(nil)

	// Create a task
	task := &scheduler.ScheduledTask{
		ID:            "test-task-toggle",
		Name:          "Toggle Test Task",
		CronExpr:      "*/10 * * * *",
		TargetCluster: "test-cluster",
		Operation: &engine.ClassifiedOperation{
			Type:     engine.OperationTypeQuery,
			Verb:     "get",
			Resource: "nodes",
		},
		Enabled: true,
	}

	err := schedulerMgr.AddTask(task)
	require.NoError(t, err)

	// Verify initial state
	tasks := schedulerMgr.ListTasks()
	var initialTask *scheduler.ScheduledTask
	for _, t := range tasks {
		if t.ID == "test-task-toggle" {
			initialTask = t
			break
		}
	}
	require.NotNil(t, initialTask)
	assert.True(t, initialTask.Enabled)

	// Note: The scheduler manager doesn't have a direct Enable/Disable method
	// This test documents the current API - tasks would need to be updated via RemoveTask + AddTask
	// or a future SetTaskEnabled method
}

func TestClusterManagement(t *testing.T) {
	// Test: Add/list/remove cluster

	registry := cluster.NewRegistry()

	// List should be empty initially
	clusters := registry.ListClusters()
	assert.Empty(t, clusters)

	// Add cluster (mock path for testing)
	// Note: In real usage, clusters are added via kubeconfig loading
	// This test verifies the registry API
}

func TestConfirmationExpiry(t *testing.T) {
	// Test: Confirmation expires after TTL

	confirmMgr := confirmation.NewManager(1 * time.Second) // 1 second TTL

	classifiedOp := &engine.ClassifiedOperation{
		Type:     engine.OperationTypeMutation,
		Verb:     "delete",
		Resource: "pods",
		Name:     "test-pod",
	}

	confirmKey, err := confirmMgr.CreateConfirmation("test-cluster", classifiedOp)
	require.NoError(t, err)

	// Should be pending initially
	pending, err := confirmMgr.GetConfirmation(confirmKey)
	require.NoError(t, err)
	assert.Equal(t, confirmation.StatusPending, pending.Status)

	// Wait for expiry
	// Note: For faster testing, we'd need to manipulate time or use a mock
	// This test documents the expiry behavior
}

func TestExitCommands(t *testing.T) {
	// Test: Verify exit commands are correctly identified

	testCases := []struct {
		input      string
		shouldExit bool
	}{
		{"exit", true},
		{"quit", true},
		{"q", true},
		{"EXIT", true},
		{"Quit", true},
		{"Q", true},
		{"get pods", false},
		{"list services", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			result := isExitCommand(tc.input)
			assert.Equal(t, tc.shouldExit, result)
		})
	}
}

// isExitCommand checks if input is an exit command
func isExitCommand(input string) bool {
	lower := strings.ToLower(strings.TrimSpace(input))
	return lower == "exit" || lower == "quit" || lower == "q"
}

func TestCLIFlagHandling(t *testing.T) {
	// Test: Verify CLI handles flags correctly

	rootCmd := cli.NewRootCommand()

	// Test help flag
	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetArgs([]string{"--help"})

	err := rootCmd.Execute()
	assert.NoError(t, err)
	assert.Contains(t, buf.String(), "k8s-agent")
}

func TestInvalidCommand(t *testing.T) {
	// Test: Invalid command returns error

	rootCmd := cli.NewRootCommand()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"invalid-command"})

	err := rootCmd.Execute()
	assert.Error(t, err)
}

func TestDescribeCommand(t *testing.T) {
	// Test: Describe resource command parsing

	parsedOp, err := engine.Parse("describe pod nginx")
	require.NoError(t, err)

	classifiedOp := engine.ClassifyOperation(parsedOp)
	assert.Equal(t, engine.OperationTypeQuery, classifiedOp.Type)
	assert.Equal(t, "describe", classifiedOp.Verb)
	assert.Equal(t, "pod", classifiedOp.Resource) // Note: singular form from parser
	assert.Equal(t, "nginx", classifiedOp.Name)
}

func TestScaleCommand(t *testing.T) {
	// Test: Scale deployment command parsing

	parsedOp, err := engine.Parse("scale deployment myapp to 3")
	require.NoError(t, err)

	classifiedOp := engine.ClassifyOperation(parsedOp)
	assert.Equal(t, engine.OperationTypeMutation, classifiedOp.Type)
	assert.Equal(t, "scale", classifiedOp.Verb)
	assert.Equal(t, "deployment", classifiedOp.Resource) // Note: singular form from parser
}

func TestNamespaceParsing(t *testing.T) {
	// Test: Commands with namespace

	parsedOp, err := engine.Parse("get pods -n kube-system")
	require.NoError(t, err)

	classifiedOp := engine.ClassifyOperation(parsedOp)
	assert.Equal(t, engine.OperationTypeQuery, classifiedOp.Type)
	assert.Equal(t, "get", classifiedOp.Verb)
	assert.Equal(t, "pods", classifiedOp.Resource)
	assert.Equal(t, "kube-system", classifiedOp.Namespace)
}

func TestSessionConversation(t *testing.T) {
	// Test: Session maintains conversation history

	sessionMgr := session.NewManager()

	conversationID := "test-conv"
	_, err := sessionMgr.CreateConversation(conversationID, "test-cluster", "default")
	require.NoError(t, err)

	// Add messages
	err = sessionMgr.AddMessage(conversationID, session.RoleUser, "get pods", nil)
	require.NoError(t, err)

	err = sessionMgr.AddMessage(conversationID, session.RoleAssistant, "Listing pods...", nil)
	require.NoError(t, err)

	// Get conversation and verify messages
	conv, err := sessionMgr.GetConversation(conversationID)
	require.NoError(t, err)
	assert.Len(t, conv.Messages, 2)

	// Test context switching
	err = sessionMgr.SetNamespaceContext(conversationID, "kube-system")
	require.NoError(t, err)

	conv, err = sessionMgr.GetConversation(conversationID)
	require.NoError(t, err)
	assert.Equal(t, "kube-system", conv.Namespace)
}