package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/x/lockdrop/types"
)

func (k Keeper) HandleReplaceLockdropProposal(ctx sdk.Context, p *types.ReplaceLockdropProposal) error {
	return k.ReplaceDistrRecords(ctx, p.Records...)
}

func (k Keeper) HandleUpdateLockdropProposal(ctx sdk.Context, p *types.UpdateLockdropProposal) error {
	return k.UpdateDistrRecords(ctx, p.Records...)
}
