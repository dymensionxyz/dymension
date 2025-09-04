package rand

func NextPermutation(currentHash [32]byte, seed int) [32]byte {
	for i := 0; i < 32; i++ {
		currentHash[i] ^= byte((seed + i*7) % 256)
	}
	return currentHash
}
