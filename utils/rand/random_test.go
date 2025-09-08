package rand_test

import (
	"slices"
	"testing"

	"cosmossdk.io/core/header"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/utils/rand"
)

func TestUniformRandom(t *testing.T) {
	t.Skip("This test is for debugging and visualizing the distribution.")

	// Prepare hash
	hash := make([]byte, 32)
	hash[31] = 5
	ctx := sdk.Context{}.WithHeaderInfo(header.Info{Hash: hash})

	const iterations = 250

	modulo := math.NewInt(10_000)
	values := make([]uint64, 0, iterations)
	total := math.ZeroInt()

	for iteration := 0; iteration < iterations; iteration++ {
		hash := ctx.HeaderInfo().Hash
		newHash := rand.NextPermutation([32]byte(hash), iteration)
		headerInfo := ctx.HeaderInfo()
		headerInfo.Hash = newHash[:]
		ctx = ctx.WithHeaderInfo(headerInfo)

		random := rand.GenerateUniformRandomMod(ctx, modulo.BigInt())
		total = total.Add(math.NewIntFromBigIntMut(random))
		values = append(values, random.Uint64())
	}

	slices.Sort(values)
	for _, v := range values {
		println(v)
	}
	t.Log("Target mean", modulo.QuoRaw(2))
	t.Log("Actual mean", total.QuoRaw(iterations))
}

func TestExpRandom(t *testing.T) {
	t.Skip("This test is for debugging and visualizing the distribution.")

	// Prepare hash
	hash := make([]byte, 32)
	hash[31] = 5
	ctx := sdk.Context{}.WithHeaderInfo(header.Info{Hash: hash})

	const iterations = 250

	budget := math.NewInt(100_000)
	pumpNum := math.NewInt(100)
	values := make([]uint64, 0, iterations)
	total := math.ZeroInt()

	for iteration := 0; iteration < iterations; iteration++ {
		hash := ctx.HeaderInfo().Hash
		newHash := rand.NextPermutation([32]byte(hash), iteration)
		headerInfo := ctx.HeaderInfo()
		headerInfo.Hash = newHash[:]
		ctx = ctx.WithHeaderInfo(headerInfo)

		random := rand.GenerateExpRandomLambda(ctx, pumpNum.BigInt(), budget.BigInt())
		total = total.Add(math.NewIntFromBigIntMut(random))
		values = append(values, random.Uint64())
	}

	slices.Sort(values)
	for _, v := range values {
		println(v)
	}
	t.Log("Target mean", budget.Quo(pumpNum))
	t.Log("Actual mean", total.QuoRaw(iterations))
}
