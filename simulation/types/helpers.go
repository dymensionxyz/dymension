package types

import (
	"errors"
	"math/rand"

	"cosmossdk.io/math"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

// RandIntBetween returns a random integer in the range [min; max).
func RandIntBetween(r *rand.Rand, min, max math.Int) (math.Int, error) {
	if min.GT(max) {
		return math.Int{}, errors.New("min cannot be greater than max")
	}

	// Calculate the range size:
	// rangeSize = max - min
	rangeSize := max.Sub(min)

	// Get a random number in the range (0; rangeSize]
	randInRange, err := simtypes.RandPositiveInt(r, rangeSize)
	if err != nil {
		return math.Int{}, err
	}

	// Adjust the random number to be in the range [min; max):
	// (0; rangeSize] + min - 1 = (min-1; max-1] = [min; max)
	return min.Add(randInRange).Sub(math.OneInt()), nil
}
