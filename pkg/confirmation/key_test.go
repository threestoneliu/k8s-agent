package confirmation

import (
	"strings"
	"testing"
)

func TestGenerateSecureConfirmKey(t *testing.T) {
	key1, err := GenerateSecureConfirmKey()
	if err != nil {
		t.Fatalf("GenerateSecureConfirmKey() error = %v", err)
	}

	if len(key1) != 6 {
		t.Errorf("GenerateSecureConfirmKey() key length = %d, want 6", len(key1))
	}

	// Verify all characters are digits
	for _, c := range key1 {
		if c < '0' || c > '9' {
			t.Errorf("GenerateSecureConfirmKey() key = %s, contains non-digit character %c", key1, c)
		}
	}

	// Verify uniqueness
	key2, err := GenerateSecureConfirmKey()
	if err != nil {
		t.Fatalf("GenerateSecureConfirmKey() second call error = %v", err)
	}
	if key1 == key2 {
		t.Errorf("GenerateSecureConfirmKey() generated duplicate keys: %s and %s", key1, key2)
	}
}

func TestGenerateSecureConfirmKey_Uniqueness(t *testing.T) {
	keys := make(map[string]bool)
	for i := 0; i < 100; i++ {
		key, err := GenerateSecureConfirmKey()
		if err != nil {
			t.Fatalf("GenerateSecureConfirmKey() iteration %d error = %v", i, err)
		}
		if keys[key] {
			t.Errorf("GenerateSecureConfirmKey() generated duplicate key: %s", key)
		}
		keys[key] = true
	}
}

func TestValidateConfirmKey(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		valid bool
	}{
		{"valid 6 digit key", "123456", true},
		{"valid 000000", "000000", true},
		{"valid 999999", "999999", true},
		{"too short", "12345", false},
		{"too long", "1234567", false},
		{"contains letter", "12345a", false},
		{"contains space", "123 56", false},
		{"empty string", "", false},
		{"whitespace only", "   ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ValidateConfirmKey(tt.key); got != tt.valid {
				t.Errorf("ValidateConfirmKey(%q) = %v, want %v", tt.key, got, tt.valid)
			}
		})
	}
}

func TestValidateConfirmKey_LengthVariations(t *testing.T) {
	// Test that only exactly 6 characters is valid
	for length := 4; length <= 8; length++ {
		key := strings.Repeat("1", length)
		expectedValid := length == 6
		if got := ValidateConfirmKey(key); got != expectedValid {
			t.Errorf("ValidateConfirmKey(%q) = %v, want %v", key, got, expectedValid)
		}
	}
}
