package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

// ImportGenesis imports the sponsorship module's state from a provided genesis state.
func (k Keeper) ImportGenesis(ctx sdk.Context, genState types.GenesisState) error {
	err := k.SetParams(ctx, genState.Params)
	if err != nil {
		return fmt.Errorf("can't set module params: %w", err)
	}

	distr := types.NewDistribution()
	for _, i := range genState.VoterInfos {
		voterAddr, errX := sdk.AccAddressFromBech32(i.Voter)
		if errX != nil {
			return fmt.Errorf("can't get delegator address from bech32 '%s': %w", i.Voter, errX)
		}

		for _, v := range i.Validators {
			valAddr, err := sdk.ValAddressFromBech32(v.Validator)
			if err != nil {
				return fmt.Errorf("can't get validator address from bech32 '%s': %w", v.Validator, err)
			}

			err = k.SaveDelegatorValidatorPower(ctx, voterAddr, valAddr, v.Power)
			if err != nil {
				return fmt.Errorf("failed to save voting power: %w", err)
			}
		}

		err := k.SaveVote(ctx, voterAddr, i.Vote)
		if err != nil {
			return fmt.Errorf("failed to save vote for voter '%s': %w", voterAddr, err)
		}

		distr = distr.Merge(i.Vote.ToDistribution())
	}

	err = k.SaveDistribution(ctx, distr)
	if err != nil {
		return fmt.Errorf("failed to save distribution: %w", err)
	}

	return nil
}

// ExportGenesis returns the sponsorship module's exported genesis.
func (k Keeper) ExportGenesis(ctx sdk.Context) (types.GenesisState, error) {
	var infos []types.VoterInfo

	const Break = true
	const Continue = false

	err := k.IterateVotes(ctx, func(voterAddr sdk.AccAddress, vote types.Vote) (bool, error) {
		var vals []types.ValidatorVotingPower
		err := k.IterateDelegatorValidatorPower(ctx, voterAddr, func(valAddr sdk.ValAddress, power math.Int) (bool, error) {
			vals = append(vals, types.ValidatorVotingPower{
				Validator: valAddr.String(),
				Power:     power,
			})
			return Continue, nil
		})
		if err != nil {
			return Break, err
		}

		infos = append(infos, types.VoterInfo{
			Voter:      voterAddr.String(),
			Vote:       vote,
			Validators: vals,
		})

		return Continue, nil
	})
	if err != nil {
		return types.GenesisState{}, fmt.Errorf("failed to iterate votes and voting powers: %w", err)
	}

	params, err := k.GetParams(ctx)
	if err != nil {
		return types.GenesisState{}, fmt.Errorf("failed to get module params: %w", err)
	}

	return types.GenesisState{
		Params:     params,
		VoterInfos: infos,
	}, nil
}
