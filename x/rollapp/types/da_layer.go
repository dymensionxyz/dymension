package types

import (
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// DALayer defines a generic DA layer from the point of view of the
// dymension settlement layer.
type DALayer interface {
	// OnRollappStateUpdate is used to verify the correctness of the provided commitment.
	// It is up to the DA Layer to decide the verification level.
	// For example, a DA layer using IBC can immediately verify inclusion
	// of the blob.
	OnRollappStateUpdate(ctx sdk.Context, commitment *codectypes.Any) error
	// VerifyMembership is called by x/rollapp during the dispute process to verify
	// the inclusion or non inclusion of the blob, or verify that it was not tampered
	// with.
	// It also provides arbitrary proof bytes that can be used to confirm the membership.
	VerifyMembership(ctx sdk.Context, commitment *codectypes.Any, proof []byte) error
}
