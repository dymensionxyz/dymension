package rand

import (
	"math"
	"math/big"
	"math/rand/v2"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/osmomath"
)

// GenerateUniformRandom draws a uniform random 64-bit variable
func GenerateUniformRandom(ctx sdk.Context) *big.Int {
	// Take block hash. Tendermint uses SHA-256.
	seed := ctx.HeaderInfo().Hash
	//nolint:gosec // math/rand/v2 is used for a purpose: we need a custom seed
	generator := rand.NewChaCha8([32]byte(seed))
	return new(big.Int).SetUint64(generator.Uint64())
}

var BigMaxUint64 = osmomath.NewIntFromUint64(math.MaxUint64)

// GenerateUniformRandomMod draws a uniform random variable in [0; modulo)
func GenerateUniformRandomMod(ctx sdk.Context, modulo *big.Int) *big.Int {
	// Generate a uniform random variable in [0; MaxUint64]
	u := GenerateUniformRandom(ctx)
	// Normalize it to [0; 1) = uniform / (MAX_UINT64 + 1)
	un := osmomath.NewDecFromBigInt(u).QuoInt(BigMaxUint64.AddRaw(1))
	// Scale it up to the modulo
	return un.MulInt(osmomath.NewIntFromBigInt(modulo)).TruncateInt().BigInt()
}

// GenerateExpRandomLambda draws an exp random variable in [0, +inf)
// with lambda = numerator / denominator and mean = denominator / numerator.
func GenerateExpRandomLambda(ctx sdk.Context, lambdaNumerator, lambdaDenominator *big.Int) *big.Int {
	// Generate a uniform random variable in [0; MaxUint64]
	u := GenerateUniformRandom(ctx)

	// Handle u = 0 case to avoid ln(0)
	// The variable is in (0; MaxUint64]
	if u.Cmp(big.NewInt(0)) == 0 {
		u = big.NewInt(1)
	}

	// Normalize it to (0; 1] = uniform / MaxUint64
	un := osmomath.NewDecFromBigInt(u).QuoInt(BigMaxUint64)
	num := osmomath.NewDecFromBigInt(lambdaNumerator)
	den := osmomath.NewDecFromBigInt(lambdaDenominator)
	// Use the inverse transform method:
	// exp = -ln(uniform) / λ = -ln(uniform) * den / num
	return un.Ln().Mul(den).Quo(num).Neg().TruncateInt().BigInt()
}
