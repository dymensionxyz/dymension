package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// ChannelKeeper defines the expected IBC channel keeper
type ChannelKeeper interface {
	LookupModuleByChannel(ctx sdk.Context, portID, channelID string) (string, *capabilitytypes.Capability, error)
}

type RollappKeeper interface {
	GetParams(ctx sdk.Context) rollapptypes.Params
	GetStateInfo(ctx sdk.Context, rollappId string, index uint64) (val rollapptypes.StateInfo, found bool)
	MustGetStateInfo(ctx sdk.Context, rollappId string, index uint64) rollapptypes.StateInfo
	GetLatestFinalizedStateIndex(ctx sdk.Context, rollappId string) (val types.StateInfoIndex, found bool)
	GetAllRollapps(ctx sdk.Context) (list []types.Rollapp)
	GetValidTransfer(
		ctx sdk.Context,
		packetData []byte,
		raPortOnHub, raChanOnHub string,
	) (data types.TransferData, err error)
}

type EIBCKeeper interface {
	EIBCDemandOrderHandler(ctx sdk.Context, rollappPacket commontypes.RollappPacket, data transfertypes.FungibleTokenPacketData) error
}
