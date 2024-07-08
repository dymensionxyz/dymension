package v3

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

type RollappKeeper interface {
	SetParams(ctx sdk.Context, params rollapptypes.Params)
	GetParams(ctx sdk.Context) rollapptypes.Params
	SetRollapp(ctx sdk.Context, rollapp rollapptypes.Rollapp)
	RemoveRollapp(ctx sdk.Context, rollappId string)
}
