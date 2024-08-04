package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

// EndBlocker TODO: delete?
func (k Keeper) EndBlocker(ctx sdk.Context) error {
	inactiveVoters := k.DequeueInactiveVoters(ctx)

	dirst := types.NewDistribution()
	for _, voter := range inactiveVoters {
		vote, err := k.GetVote(ctx, voter)
		if err != nil {
			return fmt.Errorf("could not get vote for voter %s: %w", voter, err)
		}
		minVP := k.GetParams(ctx).MinVotingPower
		if vote.VotingPower.LT(minVP) {
			dirst, err = k.revokeVote(ctx, voter, vote)
			if err != nil {
				return fmt.Errorf("could not revoke vote for voter %s: %w", voter, err)
			}
		}
	}

	err := k.SaveDistribution(ctx, dirst)
	if err != nil {
		return fmt.Errorf("could not save distribution: %w", err)
	}

	return nil
}
