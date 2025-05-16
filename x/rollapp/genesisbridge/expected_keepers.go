package genesisbridge

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v8/modules/core/04-channel/types"

	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

type RollappKeeper interface {
	RollappKeeperMinimal
	SetRollapp(ctx sdk.Context, rollapp rollapptypes.Rollapp)
	GetHooks() rollapptypes.MultiRollappHooks
	GetLatestHeight(ctx sdk.Context, rollappId string) (uint64, bool)
}

type RollappKeeperMinimal interface {
	GetRollappByPortChan(ctx sdk.Context, portID, channelID string) (*rollapptypes.Rollapp, error)
	IsCanonicalChannel(ctx sdk.Context, rollappId, portID, channelID string) bool
}

type DenomMetadataKeeper interface {
	CreateDenomMetadata(ctx sdk.Context, metadata banktypes.Metadata) error
	HasDenomMetadata(ctx sdk.Context, base string) bool
}

type TransferKeeper interface {
	SetDenomTrace(ctx sdk.Context, denomTrace transfertypes.DenomTrace)
	OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, data transfertypes.FungibleTokenPacketData) error
}
