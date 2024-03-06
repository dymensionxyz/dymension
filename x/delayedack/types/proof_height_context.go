package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
)

const (
	// userKey to store the user information inside of a context
	proofHeightCtxKey = "ibc_proof_height"
)

func NewIBCProofContext(ctx sdk.Context, sequence uint64, height clienttypes.Height) sdk.Context {
	key := fmt.Sprintf("%s_%d", proofHeightCtxKey, sequence)
	return ctx.WithValue(key, height)
}

func FromIBCProofContext(ctx sdk.Context, sequence uint64) (clienttypes.Height, bool) {
	key := fmt.Sprintf("%s_%d", proofHeightCtxKey, sequence)
	u, ok := ctx.Value(key).(clienttypes.Height)
	return u, ok
}
