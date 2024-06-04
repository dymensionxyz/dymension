package keeper

import (
	sdktypes "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) MarkGenesisAsHappened(ctx sdktypes.Context, channelID, rollappID string) error {
	rollapp, found := k.GetRollapp(ctx, rollappID)
	if !found {
		panic("expected to find rollapp")
	}

	// Validate it hasn't been triggered yet
	if rollapp.GenesisState.GenesisEventHappened {
		k.Logger(ctx).Error("genesis event already happened")
		// panic(errors.New("genesis event already happened - it shouldn't have")) TODO:
	}

	rollapp.GenesisState.GenesisEventHappened = true
	k.SetRollapp(ctx, rollapp)

	return nil
}
