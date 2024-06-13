package types

import (
	tmbytes "github.com/cometbft/cometbft/libs/bytes"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// TransferKeeper defines the expected transfer keeper
type TransferKeeper interface {
	HasDenomTrace(ctx sdk.Context, denomTraceHash tmbytes.HexBytes) bool
}

// ChannelKeeper defines the expected IBC channel keeper
type ChannelKeeper interface {
	GetChannelClientState(ctx sdk.Context, portID, channelID string) (string, exported.ClientState, error)
}

// BankKeeper defines the expected interface needed
type BankKeeper interface {
	HasDenomMetaData(ctx sdk.Context, denom string) bool
	SetDenomMetaData(ctx sdk.Context, denomMetaData types.Metadata)
}

type RollappKeeper interface {
	GetParams(ctx sdk.Context) rollapptypes.Params
	GetRollapp(ctx sdk.Context, chainID string) (rollapp rollapptypes.Rollapp, found bool)
	RegisterDenomMetadata(ctx sdk.Context, rollapp rollapptypes.Rollapp, ibcBaseDenom, baseDenom string)
}
