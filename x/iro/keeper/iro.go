package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/iro/types"
)

// SetPlan sets a specific plan in the store from its index
func (k Keeper) SetPlan(ctx sdk.Context, plan types.Plan) {
	store := ctx.KVStore(k.storeKey)
	b := k.cdc.MustMarshal(&plan)
	store.Set(types.PlanKey(fmt.Sprintf("%d", plan.Id)), b)

	planByRollappKey := types.PlansByRollappKey(plan.RollappId)
	// Store the plan ID instead of the plan itself
	store.Set(planByRollappKey, []byte(fmt.Sprintf("%d", plan.Id)))
}

// GetPlan returns a plan from its index
func (k Keeper) GetPlan(ctx sdk.Context, planId string) (val types.Plan, found bool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.PlanKey(planId))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// GetPlanByRollapp returns a plan from its rollapp ID
func (k Keeper) GetPlanByRollapp(ctx sdk.Context, rollappId string) (val types.Plan, found bool) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.PlansByRollappKey(rollappId))
	if b == nil {
		return val, false
	}

	planId := string(b)
	return k.GetPlan(ctx, planId)
}

// MustGetPlan returns a plan from its index
// It will panic if the plan is not found
func (k Keeper) MustGetPlan(ctx sdk.Context, planId string) types.Plan {
	plan, found := k.GetPlan(ctx, planId)
	if !found {
		panic(fmt.Sprintf("plan not found for ID: %s", planId))
	}
	return plan
}

// MustGetPlanByRollapp returns a plan from its rollapp ID
// It will panic if the plan is not found
func (k Keeper) MustGetPlanByRollapp(ctx sdk.Context, rollappId string) types.Plan {
	plan, found := k.GetPlanByRollapp(ctx, rollappId)
	if !found {
		panic(fmt.Sprintf("plan not found for rollapp ID: %s", rollappId))
	}
	return plan
}

// GetAllPlans returns all plans
func (k Keeper) GetAllPlans(ctx sdk.Context) (list []types.Plan) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PlanKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close() // nolint: errcheck

	for ; iterator.Valid(); iterator.Next() {
		var val types.Plan
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// SetLastPlanId sets the last plan ID in the store
func (k Keeper) SetLastPlanId(ctx sdk.Context, lastPlanId uint64) {
	store := ctx.KVStore(k.storeKey)
	b := sdk.Uint64ToBigEndian(lastPlanId)
	store.Set(types.LastPlanIdKey, b)
}

// GetLastPlanId returns the last plan ID from the store
func (k Keeper) GetLastPlanId(ctx sdk.Context) (lastPlanId uint64) {
	store := ctx.KVStore(k.storeKey)
	b := store.Get(types.LastPlanIdKey)
	if b == nil {
		return 0
	}

	return sdk.BigEndianToUint64(b)
}

func (k Keeper) GetNextPlanIdAndIncrement(ctx sdk.Context) uint64 {
	lastPlanId := k.GetLastPlanId(ctx)
	k.SetLastPlanId(ctx, lastPlanId+1)
	return lastPlanId + 1
}
