package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	eibctypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

var _ eibctypes.EIBCHooks = eibcHooks{}

type eibcHooks struct {
	eibctypes.BaseEIBCHook
	Keeper
}

func (k Keeper) GetEIBCHooks() eibctypes.EIBCHooks {
	return eibcHooks{
		BaseEIBCHook: eibctypes.BaseEIBCHook{},
		Keeper:       k,
	}
}

// AfterDemandOrderFulfilled is called every time a demand order is fulfilled.
// Once it is fulfilled the underlying packet recipient should be updated to the fullfiller.
func (k eibcHooks) AfterDemandOrderFulfilled(ctx sdk.Context, demandOrder *eibctypes.DemandOrder, fulfillerAddress string) error {
	err := k.UpdateRollappPacketRecipient(ctx, demandOrder.TrackingPacketKey, fulfillerAddress)
	if err != nil {
		return err
	}
	return nil
}
