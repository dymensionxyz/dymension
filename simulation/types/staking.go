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

	srcAddr := sdk.MustAccAddressFromBech32(srcVal.GetOperator())
	delegations, err := k.GetValidatorDelegations(ctx, sdk.ValAddress(srcAddr))
	if delegations == nil || err != nil {
		return nil
	}

	return &delegations[r.Intn(len(delegations))]
}
