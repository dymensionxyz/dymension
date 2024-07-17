package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const InterchainDALayerName = "interchain"

var _ DALayer = InterchainDALayer{}

type InterchainDALayer struct{}

func NewInterchainDALayer() InterchainDALayer {
	return InterchainDALayer{}
}

// OnRollappStateUpdate implements DALayer.
func (da InterchainDALayer) OnRollappStateUpdate(ctx sdk.Context, commitment *codectypes.Any) error {
	return nil // panic("unimplemented 👻")
}

// VerifyMembership implements DALayer.
func (da InterchainDALayer) VerifyMembership(ctx sdk.Context, commitment *codectypes.Any, proof []byte) error {
	return nil // panic("unimplemented 👻")
}
