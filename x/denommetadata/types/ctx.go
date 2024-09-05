package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type denomMetaRollappIDKey struct{}

func CtxWithRollappID(ctx sdk.Context, rollappID string) sdk.Context {
	return ctx.WithValue(denomMetaRollappIDKey{}, rollappID)
}

func CtxToRollappID(ctx sdk.Context) (string, bool) {
	id := ctx.Value(denomMetaRollappIDKey{})
	if id == nil {
		return "", false
	}
	idS, ok := id.(string)
	return idS, ok
}
