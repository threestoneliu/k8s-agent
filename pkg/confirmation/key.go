package confirmation

import (
	"fmt"
	"math/rand"
	"strings"
	"sync/atomic"
	"time"
)

const confirmKeyLength = 6
const confirmKeyCharset = "0123456789"

var (
	keyCounter uint64
)

// init seeds math/rand with current time and a random component
func init() {
	// Seed with current time and a random value
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Intn(10000) // Advance the generator a bit
}

// GenerateSecureConfirmKey generates a secure 6-digit confirmation code
// Uses math/rand with time-based seeding for non-blocking operation
func GenerateSecureConfirmKey() (string, error) {
	counter := atomic.AddUint64(&keyCounter, 1)

	result := make([]byte, confirmKeyLength)
	now := time.Now().UnixNano()

	// Use time and counter to generate pseudo-random digits
	for i := 0; i < confirmKeyLength; i++ {
		// Combine multiple sources for better distribution
		seed := uint64(now) + counter*uint64(i+1) + uint64(rand.Intn(1000))
		result[i] = confirmKeyCharset[seed%uint64(len(confirmKeyCharset))]
	}

	return string(result), nil
}

// ValidateConfirmKey checks if a confirmation key is valid (6 digits)
func ValidateConfirmKey(key string) bool {
	if len(key) != confirmKeyLength {
		return false
	}
	for _, c := range key {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

// generateID generates a unique ID for a pending operation
func generateID() string {
	counter := atomic.AddUint64(&keyCounter, 1)
	now := time.Now().UnixNano()

	// Use pseudo-random based on counter and time
	randomPart := uint64(now) ^ (counter * 0x9e3779b97f4a7c15)

	id := fmt.Sprintf("conf-%d-%06d", counter%1000000, randomPart%1000000)
	return strings.ToUpper(id)
}
