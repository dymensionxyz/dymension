package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/dymensionxyz/dymension/v3/x/sponsorship/types"
)

func (k Keeper) Vote(ctx sdk.Context, voter sdk.AccAddress, weights []types.GaugeWeight) (types.Vote, error) {
	params := k.GetParams(ctx)

	// Check all gauges have sufficient weights and are perpetual
	for _, weight := range weights {
		if weight.Weight.LT(params.MinAllocationWeight) {
			return types.Vote{}, fmt.Errorf("gauge weight '%d' is less than min allocation weight '%d'", weight.Weight.Int64(), params.MinAllocationWeight.Int64())
		}

		gauge, err := k.incentivesKeeper.GetGaugeByID(ctx, weight.GaugeId)
		if err != nil {
			return types.Vote{}, fmt.Errorf("failed to get gauge by id '%d': %w", weight.GaugeId, err)
		}

		if !gauge.IsPerpetual {
			return types.Vote{}, fmt.Errorf("gauge '%d' is not perpetual", weight.GaugeId)
		}
	}

	return types.Vote{}, nil
}

func (k Keeper) RevokeVote(ctx sdk.Context, voter sdk.AccAddress) error {
	panic("not implemented")
}

func (k Keeper) UpdateVotingPower(ctx sdk.Context, voter sdk.AccAddress, power math.Int) error {
	panic("not implemented")
}

func (k Keeper) GetUserVotingPower(ctx sdk.Context, voter sdk.AccAddress) {
	k.stakingKeeper.IterateBondedValidatorsByPower(ctx, func(index int64, validator stakingtypes.ValidatorI) (stop bool) {

	})
	validators := k.stakingKeeper.IterateDelegations(ctx, voter, func(index int64, del stakingtypes.DelegationI) (stop bool) {
		validator, err := k.stakingKeeper.GetDelegatorValidator(ctx, del.GetDelegatorAddr(), del.GetValidatorAddr())
		validator.IsBonded()
	})
}
