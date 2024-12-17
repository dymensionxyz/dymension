package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type EIBCHooks interface {
	AfterDemandOrderFulfilled(ctx sdk.Context, demandOrder *DemandOrder, receiverAddr string) error
}

type MultiEIBCHooks []EIBCHooks

var _ EIBCHooks = MultiEIBCHooks{}

func NewMultiEIBCHooks(hooks ...EIBCHooks) MultiEIBCHooks {
	return hooks
}

func (h MultiEIBCHooks) AfterDemandOrderFulfilled(ctx sdk.Context, demandOrder *DemandOrder, receiverAddr string) error {
	for i := range h {
		err := h[i].AfterDemandOrderFulfilled(ctx, demandOrder, receiverAddr)
		if err != nil {
			return err
		}
	}
	return nil
}

type BaseEIBCHook struct{}

var _ EIBCHooks = BaseEIBCHook{}

func (b BaseEIBCHook) AfterDemandOrderFulfilled(ctx sdk.Context, demandOrder *DemandOrder, fulfillerAddress string) error {
	return nil
}
