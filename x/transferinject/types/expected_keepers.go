package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	delayedacktypes "github.com/dymensionxyz/dymension/v3/x/delayedack/types"

	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// BankKeeper defines the expected interface needed
type BankKeeper interface {
	GetDenomMetaData(ctx sdk.Context, denom string) (types.Metadata, bool)
}

type RollappKeeper interface {
	SetRollapp(ctx sdk.Context, rollapp rollapptypes.Rollapp)
	MustGetRollapp(ctx sdk.Context, rollappId string) rollapptypes.Rollapp
	GetValidTransfer(
		ctx sdk.Context,
		packetData []byte,
		raPortOnHub, raChanOnHub string,
	) (data delayedacktypes.TransferData, err error)
}
