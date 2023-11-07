package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authzcodec "github.com/cosmos/cosmos-sdk/x/authz/codec"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

// RegisterCodec registers the necessary x/streamer interfaces and concrete types on the provided
// LegacyAmino codec. These types are used for Amino JSON serialization.
func RegisterCodec(cdc *codec.LegacyAmino) {
}

// RegisterInterfaces registers interfaces and implementations of the streamer module.
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*govtypes.Content)(nil),
		&CreateStreamProposal{},
	)
}

func init() {
	RegisterCodec(amino)
	// Register all Amino interfaces and concrete types on the authz Amino codec so that this can later be
	// used to properly serialize MsgGrant and MsgExec instances
	sdk.RegisterLegacyAminoCodec(amino)
	RegisterCodec(authzcodec.Amino)

	amino.Seal()
}
