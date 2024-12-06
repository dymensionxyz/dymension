package types

import (
	"errors"
	"math/rand"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

// RandIntBetween returns a random integer in the range [min; max).
// There is a similar method for 'int' in sdk sim utils which is [min,max]
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

func RandChoice[T any](r *rand.Rand, choices []T) T {
	return choices[r.Intn(len(choices))]
}

func RandFutureTime(r *rand.Rand, ctx sdk.Context, maxDuration time.Duration) time.Time {
	return ctx.BlockTime().Add(RandDuration(r, maxDuration))
}

func RandDuration(r *rand.Rand, maxDuration time.Duration) time.Duration {
	return time.Duration(r.Int63n(int64(maxDuration)))
}

func AccByBech32(accs []simtypes.Account, address string) simtypes.Account {
	return AccByAddr(accs, sdk.MustAccAddressFromBech32(address))
}

func AccByAddr(accs []simtypes.Account, address sdk.AccAddress) simtypes.Account {
	ret, ok := simtypes.FindAccount(accs, address)
	if !ok {
		panic("acc by addr")
	}
	return ret
	//for _, acc := range accs {
	//	if acc.Address.Equals(address) {
	//		return acc
	//	}
	//}
	//panic("acc by addr")
}
