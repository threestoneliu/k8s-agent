package cli

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRootCommand(t *testing.T) {
	cmd := NewRootCommand()

	assert.NotNil(t, cmd)
	assert.Equal(t, "k8s-agent", cmd.Name())
	assert.True(t, cmd.HasSubCommands()) // Should have subcommands
}

func TestRootCommandSubCommands(t *testing.T) {
	cmd := NewRootCommand()

	// Get all subcommands
	subCommands := cmd.Commands()
	subCommandNames := make([]string, len(subCommands))
	for i, sub := range subCommands {
		subCommandNames[i] = sub.Name()
	}

	// Verify expected subcommands exist
	expectedCommands := []string{"chat", "cluster"}
	for _, expected := range expectedCommands {
		found := false
		for _, name := range subCommandNames {
			if name == expected {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected subcommand %q not found", expected)
	}
}

func TestRootCommandExecution(t *testing.T) {
	cmd := NewRootCommand()

	// Test execution with no args - should show help
	assert.NotPanics(t, func() {
		buf := &bytes.Buffer{}
		cmd.SetOut(buf)
		cmd.SetErr(buf)
		cmd.SetArgs([]string{})
		err := cmd.Execute()
		// Should not error for empty args
		assert.NoError(t, err)
	})
}

func TestChatCommand(t *testing.T) {
	cmd := NewRootCommand()
	chatCmd, _, err := cmd.Find([]string{"chat"})
	require.NoError(t, err)
	assert.Equal(t, "chat", chatCmd.Name())
	assert.NotNil(t, chatCmd.RunE)
}

func TestClusterCommand(t *testing.T) {
	cmd := NewRootCommand()
	clusterCmd, _, err := cmd.Find([]string{"cluster"})
	require.NoError(t, err)
	assert.Equal(t, "cluster", clusterCmd.Name())

	// Cluster should have subcommands
	subCommands := clusterCmd.Commands()
	subCommandNames := make([]string, len(subCommands))
	for i, sub := range subCommands {
		subCommandNames[i] = sub.Name()
	}

	expectedClusterCommands := []string{"list", "add", "use", "remove"}
	for _, expected := range expectedClusterCommands {
		found := false
		for _, name := range subCommandNames {
			if name == expected {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected cluster subcommand %q not found", expected)
	}
}

