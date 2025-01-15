package urand

import (
	"fmt"

	"github.com/cometbft/cometbft/libs/rand"
)

// RollappID generates a unique rollapp ID, following the pattern: "name_1234-1"
func RollappID() string {
	return fmt.Sprintf("%s_%d-1", RollappAlias(), rand.Int63())
}

func RollappAlias() string {
	alias := make([]byte, 8)
	for i := range alias {
		alias[i] = byte(rand.Intn('z'-'a'+1) + 'a')
	}
	return string(alias)
}
