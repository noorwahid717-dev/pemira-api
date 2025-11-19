package voting

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// generateTokenHash generates a secure random token hash
func generateTokenHash(electionID, voterID int64) string {
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		// Fallback to timestamp-based if random fails
		return fmt.Sprintf("vt_%d_%d_%d", electionID, voterID, generateFallbackRandom())
	}
	
	// Combine random bytes with election and voter IDs for uniqueness
	data := append(randomBytes, []byte(fmt.Sprintf("%d:%d", electionID, voterID))...)
	hash := sha256.Sum256(data)
	
	return "vt_" + hex.EncodeToString(hash[:12])
}

func generateFallbackRandom() int64 {
	b := make([]byte, 8)
	rand.Read(b)
	var n int64
	for i := 0; i < 8; i++ {
		n = n<<8 | int64(b[i])
	}
	if n < 0 {
		n = -n
	}
	return n
}
