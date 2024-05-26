package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"

	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// BankKeeper defines the expected interface needed
type BankKeeper interface {
	GetDenomMetaData(ctx sdk.Context, denom string) (types.Metadata, bool)
	HasDenomMetaData(ctx sdk.Context, denom string) bool
}

type RollappKeeper interface {
	SetRollapp(ctx sdk.Context, rollapp rollapptypes.Rollapp)
	ExtractRollappFromChannel(
		ctx sdk.Context,
		rollappPortOnHub string,
		rollappChannelOnHub string,
	) (*rollapptypes.Rollapp, error)
}
