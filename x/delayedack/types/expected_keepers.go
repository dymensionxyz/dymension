package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/ibc-go/modules/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
	eibctypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// ChannelKeeper defines the expected IBC channel keeper
type ChannelKeeper interface {
	LookupModuleByChannel(ctx sdk.Context, portID, channelID string) (string, *capabilitytypes.Capability, error)
	SetPacketCommitment(ctx sdk.Context, portID, channelID string, sequence uint64, commitmentHash []byte)
}

type RollappKeeper interface {
	MustGetStateInfo(ctx sdk.Context, rollappId string, index uint64) types.StateInfo
	GetLatestFinalizedHeight(ctx sdk.Context, rollappId string) (uint64, error)
	IsHeightFinalized(ctx sdk.Context, rollappID string, height uint64) bool
	GetAllRollapps(ctx sdk.Context) (list []types.Rollapp)
	GetValidTransfer(
		ctx sdk.Context,
		packetData []byte,
		raPortOnHub, raChanOnHub string,
	) (data types.TransferData, err error)
}

type EIBCKeeper interface {
	EIBCDemandOrderHandler(ctx sdk.Context, rollappPacket commontypes.RollappPacket, data transfertypes.FungibleTokenPacketData) error
	PendingOrderByPacket(ctx sdk.Context, p *commontypes.RollappPacket) (*eibctypes.DemandOrder, error)
}
