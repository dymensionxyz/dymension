package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

type BankKeeper interface {
	HasDenomMetaData(ctx sdk.Context, denom string) bool
	SetDenomMetaData(ctx sdk.Context, denomMetaData types.Metadata)
}

type RollappKeeper interface {
	GetParams(ctx sdk.Context) rollapptypes.Params
}
