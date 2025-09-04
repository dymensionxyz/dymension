package rand

import (
	"crypto/sha256"
	"math/big"
	"math/rand/v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GenerateUnifiedRandom draws a uniform random variable in [0; 1)
func GenerateUnifiedRandom(ctx sdk.Context) *big.Float {
	// Take block hash
	hash := ctx.HeaderInfo().Hash

	// Get 32 bytes hash from it
	h := sha256.New()
	h.Write(hash)
	seed := [32]byte(h.Sum(nil))

	// Use a PRNG to generate a float in [0; 1)
	//nolint:gosec // math/rand/v2 is used for purpose
	generator := rand.New(rand.NewChaCha8(seed))
	return big.NewFloat(generator.Float64())
}

// GenerateUnifiedRandomMod draws a uniform random variable in [0; modulo)
func GenerateUnifiedRandomMod(ctx sdk.Context, modulo *big.Int) *big.Float {
	randFloat := GenerateUnifiedRandom(ctx)
	randFloat.Mul(randFloat, new(big.Float).SetInt(modulo))
	return randFloat
}

// GenerateUnifiedRandomModInt draws a uniform random variable in [0; modulo).
// Result is truncated to Int.
// If `result` is provided, store the result there instead of allocating a new Int.
func GenerateUnifiedRandomModInt(ctx sdk.Context, modulo *big.Int, result *big.Int) *big.Int {
	randFloat := GenerateUnifiedRandomMod(ctx, modulo)
	result, _ = randFloat.Int(result)
	return result
}

// GenerateExpRandom draws an exp random variable in (0, +math.MaxFloat64]
// with lambda = 1 and mean = 1.
func GenerateExpRandom(ctx sdk.Context) *big.Float {
	// Take block hash
	hash := ctx.HeaderInfo().Hash

	// Get 32 bytes hash from it
	h := sha256.New()
	h.Write(hash)
	seed := [32]byte(h.Sum(nil))

	// Use a PRNG to generate an exp float in (0, +math.MaxFloat64]
	//nolint:gosec // math/rand/v2 is used for purpose
	generator := rand.New(rand.NewChaCha8(seed))
	return big.NewFloat(generator.ExpFloat64())
}

// GenerateExpRandomLambda draws an exp random variable in (0, +math.MaxFloat64]
// with lambda = numerator / denominator and mean = denominator / numerator.
func GenerateExpRandomLambda(ctx sdk.Context, numerator, denominator *big.Int) *big.Float {
	// scaledRnd = expRnd / lambda = expRnd * denominator / numerator
	expRnd := GenerateExpRandom(ctx)
	expRnd.Mul(expRnd, new(big.Float).SetInt(denominator))
	expRnd.Quo(expRnd, new(big.Float).SetInt(numerator))
	return expRnd
}

// GenerateExpRandomLambdaInt draws a random exp variable in (0, +math.MaxFloat64]
// with lambda = numerator / denominator and mean = denominator / numerator.
// Result is truncated to Int.
// If `result` is provided, store the result there instead of allocating a new Int.
func GenerateExpRandomLambdaInt(ctx sdk.Context, numerator, denominator *big.Int, result *big.Int) *big.Int {
	expRnd := GenerateExpRandomLambda(ctx, numerator, denominator)
	result, _ = expRnd.Int(result)
	return result
}
