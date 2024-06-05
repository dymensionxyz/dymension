package keeper

import sdktypes "github.com/cosmos/cosmos-sdk/types"

func (k Keeper) MarkGenesisAsHappened(ctx sdktypes.Context, channelID, rollappID string) error {
	rollapp, found := k.GetRollapp(ctx, rollappID)
	if !found {
		panic("expected to find rollapp")
	}

	// TODO: something with transfers enabled?

	k.SetRollapp(ctx, rollapp)

	return nil
}

func ()