package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type ctxKeySkip struct{}

func SkipContext(ctx sdk.Context) sdk.Context {
	return ctx.WithValue(ctxKeySkip{}, true)
}

func Skip(ctx sdk.Context) bool {
	val, ok := ctx.Value(ctxKeySkip{}).(bool)
	return ok && val
}
