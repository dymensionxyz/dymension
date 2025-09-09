package rand

import (
	"math/rand/v2"
)

func NextPermutation(currentHash [32]byte, seed int) [32]byte {
	// Use current hash + seed as ChaCha8 key
	var key [32]byte
	copy(key[:], currentHash[:])

	// Mix in the seed
	for i := 0; i < 4; i++ {
		key[i] ^= byte(seed >> (i * 8))
	}

	rng := rand.NewChaCha8(key)
	_, _ = rng.Read(currentHash[:])
	return currentHash
}
