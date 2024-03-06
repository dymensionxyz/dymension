package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
)

const (
	// proofHeightCtxKey is a context key to pass the proof height from the msg to the IBC middleware
	proofHeightCtxKey = "ibc_proof_height"
)

func NewIBCProofContext(ctx sdk.Context, packetId channeltypes.PacketId, height clienttypes.Height) sdk.Context {
	key := fmt.Sprintf("%s_%s", proofHeightCtxKey, packetId.String())
	return ctx.WithValue(key, height)
}

func FromIBCProofContext(ctx sdk.Context, packetId channeltypes.PacketId) (clienttypes.Height, bool) {
	key := fmt.Sprintf("%s_%s", proofHeightCtxKey, packetId.String())
	u, ok := ctx.Value(key).(clienttypes.Height)
	return u, ok
}
