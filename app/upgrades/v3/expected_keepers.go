package v3

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	delayedacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	seqtypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

type RollappKeeper interface {
	SetParams(ctx sdk.Context, params rollapptypes.Params)
}

type SequencerKeeper interface {
	SetParams(ctx sdk.Context, params seqtypes.Params)
}

type DelayedAckKeeper interface {
	SetParams(ctx sdk.Context, params delayedacktypes.Params)
}
