package transfergenesis

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	irotypes "github.com/dymensionxyz/dymension/v3/x/iro/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

type RollappKeeper interface {
	GetRollappByPortChan(ctx sdk.Context, raPortOnHub, raChanOnHub string) (*rollapptypes.Rollapp, error)
	MustGetRollapp(ctx sdk.Context, rollappId string) rollapptypes.Rollapp
	SetRollapp(ctx sdk.Context, rollapp rollapptypes.Rollapp)
	GetHooks() rollapptypes.MultiRollappHooks
}

type DenomMetadataKeeper interface {
	CreateDenomMetadata(ctx sdk.Context, metadata banktypes.Metadata) error
	HasDenomMetadata(ctx sdk.Context, base string) bool
}

type TransferKeeper interface {
	SetDenomTrace(ctx sdk.Context, denomTrace transfertypes.DenomTrace)
	OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, data transfertypes.FungibleTokenPacketData) error
}

type IROKeeper interface {
	MustGetPlanByRollapp(ctx sdk.Context, rollappID string) irotypes.Plan
	GetPlanByRollapp(ctx sdk.Context, rollappID string) (irotypes.Plan, bool)
	GetModuleAccountAddress() string
}
