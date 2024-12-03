package types

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
)

// RandomDelegation returns a random delegation from a random validator.
func RandomDelegation(ctx sdk.Context, r *rand.Rand, k StakingKeeper) *stakingtypes.Delegation {
	allVals := k.GetAllValidators(ctx)
	srcVal, ok := testutil.RandSliceElem(r, allVals)
	if !ok {
		return nil
	}

	srcAddr := srcVal.GetOperator()
	delegations := k.GetValidatorDelegations(ctx, srcAddr)
	if delegations == nil {
		return nil
	}

	return &delegations[r.Intn(len(delegations))]
}
