package types

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// RandomDelegation returns a random delegation from a random validator.
func RandomDelegation(ctx sdk.Context, r *rand.Rand, k StakingKeeper) *stakingtypes.Delegation {
	allVals, err := k.GetAllValidators(ctx)
	if err != nil {
		return nil
	}
	srcVal, ok := testutil.RandSliceElem(r, allVals)
	if !ok {
		return nil
	}

	srcValAddr, err := sdk.ValAddressFromBech32(srcVal.GetOperator())
	if err != nil {
		return nil
	}
	delegations, err := k.GetValidatorDelegations(ctx, srcValAddr)
	if delegations == nil || err != nil {
		return nil
	}

	return &delegations[r.Intn(len(delegations))]
}
