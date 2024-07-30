package keeper

import (
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/dymensionxyz/dymension/v3/app/params"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(
		k.DisputePeriodInBlocks(ctx),
		k.AliasFeeTable(ctx),
	)
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}

// DisputePeriodInBlocks returns the DisputePeriodInBlocks param
func (k Keeper) DisputePeriodInBlocks(ctx sdk.Context) (res uint64) {
	k.paramstore.Get(ctx, types.KeyDisputePeriodInBlocks, &res)
	return
}

func (k Keeper) AliasFeeTable(ctx sdk.Context) (res map[string]sdk.Coin) {
	k.paramstore.Get(ctx, types.KeyAliasFeeTable, &res)
	return
}

func (k Keeper) GetPriceForAlias(ctx sdk.Context, alias string) sdk.Coin {
	aliasPricingTable := k.AliasFeeTable(ctx)
	aliasLength := int32(len(alias))

	maxAliasLength := int32(0)
	for lengthStr := range aliasPricingTable {
		length, err := strconv.ParseInt(lengthStr, 10, 32)
		if err != nil {
			panic(err)
		}

		if int32(length) > maxAliasLength {
			maxAliasLength = int32(length)
		}
	}

	// if alias length not defined in pricing table, use the max length that is defined
	if aliasLength > maxAliasLength {
		aliasLength = maxAliasLength
	}

	if aliasPrice, ok := aliasPricingTable[fmt.Sprint(aliasLength)]; ok {
		return aliasPrice
	}

	return sdk.NewCoin(appparams.BaseDenom, sdk.NewInt(0))
}
