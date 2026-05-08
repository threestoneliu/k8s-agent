package session

import (
	"errors"
	"fmt"
	"sync"
	"testing"
)

func TestStore_NewStore(t *testing.T) {
	s := NewStore()

	if s == nil {
		t.Fatal("NewStore should not return nil")
	}
	if s.convMap == nil {
		t.Error("convMap should be initialized")
	}
}

func TestStore_Errors(t *testing.T) {
	// Ensure error variables are distinct
	errs := []error{ErrConversationNotFound, ErrConversationAlreadyExists, ErrInvalidConversationID}
	seen := make(map[error]bool)

	for _, err := range errs {
		if seen[err] {
			t.Errorf("duplicate error value detected: %v", err)
		}
		seen[err] = true
	}

	// Ensure errors are not nil
	for _, err := range errs {
		if err == nil {
			t.Error("error variable is nil")
		}
	}
}

func TestStore_CreateConversation(t *testing.T) {
	s := NewStore()

	conv, err := s.CreateConversation("test-id", "cluster-a", "default")
	if err != nil {
		t.Fatalf("CreateConversation() error = %v", err)
	}
	if conv == nil {
		t.Fatal("CreateConversation returned nil conversation")
	}
	if conv.ID != "test-id" {
		t.Errorf("ID = %v, want test-id", conv.ID)
	}
	if conv.ClusterName != "cluster-a" {
		t.Errorf("ClusterName = %v, want cluster-a", conv.ClusterName)
	}
	if conv.Namespace != "default" {
		t.Errorf("Namespace = %v, want default", conv.Namespace)
	}
}

func TestStore_CreateConversation_Duplicate(t *testing.T) {
	s := NewStore()

	_, err := s.CreateConversation("dup-id", "", "")
	if err != nil {
		t.Fatalf("First CreateConversation() failed: %v", err)
	}

	_, err = s.CreateConversation("dup-id", "", "")
	if !errors.Is(err, ErrConversationAlreadyExists) {
		t.Errorf("Duplicate error = %v, want ErrConversationAlreadyExists", err)
	}
}

func TestStore_CreateConversation_EmptyID(t *testing.T) {
	s := NewStore()

	_, err := s.CreateConversation("", "", "")
	if !errors.Is(err, ErrInvalidConversationID) {
		t.Errorf("Empty ID error = %v, want ErrInvalidConversationID", err)
	}
}

func TestStore_GetConversation(t *testing.T) {
	s := NewStore()

	// Non-existent
	_, err := s.GetConversation("non-existent")
	if !errors.Is(err, ErrConversationNotFound) {
		t.Errorf("GetConversation() = %v, want ErrConversationNotFound", err)
	}

	// Existing
	created, _ := s.CreateConversation("get-test", "cluster", "ns")
	retrieved, err := s.GetConversation("get-test")
	if err != nil {
		t.Fatalf("GetConversation() error = %v", err)
	}
	if retrieved.ID != created.ID {
		t.Errorf("Retrieved ID = %v, want %v", retrieved.ID, created.ID)
	}
}

func TestStore_DeleteConversation(t *testing.T) {
	s := NewStore()

	// Delete non-existent
	err := s.DeleteConversation("non-existent")
	if !errors.Is(err, ErrConversationNotFound) {
		t.Errorf("DeleteConversation() = %v, want ErrConversationNotFound", err)
	}

	// Delete existing
	s.CreateConversation("delete-test", "", "")
	err = s.DeleteConversation("delete-test")
	if err != nil {
		t.Fatalf("DeleteConversation() error = %v", err)
	}

	// Verify deleted
	_, err = s.GetConversation("delete-test")
	if !errors.Is(err, ErrConversationNotFound) {
		t.Errorf("After delete, GetConversation() = %v, want ErrConversationNotFound", err)
	}
}

func TestStore_ListConversations(t *testing.T) {
	s := NewStore()

	// Empty list
	convs := s.ListConversations()
	if len(convs) != 0 {
		t.Errorf("Empty store, ListConversations() = %d, want 0", len(convs))
	}

	// Add some
	s.CreateConversation("conv-1", "cluster-a", "ns-1")
	s.CreateConversation("conv-2", "cluster-b", "ns-2")

	convs = s.ListConversations()
	if len(convs) != 2 {
		t.Errorf("ListConversations() = %d, want 2", len(convs))
	}
}

func TestStore_ListConversations_ReturnsCopy(t *testing.T) {
	s := NewStore()
	s.CreateConversation("conv-1", "", "")

	list1 := s.ListConversations()
	list2 := s.ListConversations()

	if &list1 == &list2 {
		t.Error("ListConversations() should return a new slice each time")
	}
}

func TestStore_UpdateConversation(t *testing.T) {
	s := NewStore()

	// Update non-existent
	_, err := s.UpdateConversation("non-existent", func(c *Conversation) error {
		return nil
	})
	if !errors.Is(err, ErrConversationNotFound) {
		t.Errorf("UpdateConversation() = %v, want ErrConversationNotFound", err)
	}

	// Update existing
	s.CreateConversation("update-test", "", "")
	updated, err := s.UpdateConversation("update-test", func(c *Conversation) error {
		c.SetClusterContext("new-cluster")
		return nil
	})
	if err != nil {
		t.Fatalf("UpdateConversation() error = %v", err)
	}
	if updated.GetClusterContext() != "new-cluster" {
		t.Errorf("Cluster context = %v, want new-cluster", updated.GetClusterContext())
	}
}

func TestStore_UpdateConversation_WithError(t *testing.T) {
	s := NewStore()
	s.CreateConversation("error-test", "", "")

	expectedErr := errors.New("update failed")
	_, err := s.UpdateConversation("error-test", func(c *Conversation) error {
		return expectedErr
	})
	if !errors.Is(err, expectedErr) {
		t.Errorf("UpdateConversation() error = %v, want %v", err, expectedErr)
	}
}

func TestStore_StoreConcurrency(t *testing.T) {
	s := NewStore()
	var wg sync.WaitGroup

	// Concurrent creates
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			s.CreateConversation(fmt.Sprintf("cluster-%d", id), "", "")
		}(i)
	}

	wg.Wait()

	// All should be present
	count := len(s.ListConversations())
	if count != 100 {
		t.Errorf("After concurrent creates, count = %d, want 100", count)
	}
}

// Benchmark tests
func BenchmarkStore_CreateConversation(b *testing.B) {
	s := NewStore()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.CreateConversation(fmt.Sprintf("cluster-%d", i), "", "")
	}
}

func BenchmarkStore_GetConversation(b *testing.B) {
	s := NewStore()
	for i := 0; i < 100; i++ {
		s.CreateConversation(fmt.Sprintf("cluster-%d", i), "", "")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = s.GetConversation("cluster-0")
	}
}
