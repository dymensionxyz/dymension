package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type EIBCHooks interface {
	AfterDemandOrderFulfilled(ctx sdk.Context, demandOrder *DemandOrder, newTransferRecipient string) error
}

type MultiEIBCHooks []EIBCHooks

var _ EIBCHooks = MultiEIBCHooks{}

func NewMultiEIBCHooks(hooks ...EIBCHooks) MultiEIBCHooks {
	return hooks
}

func (h MultiEIBCHooks) AfterDemandOrderFulfilled(ctx sdk.Context, o *DemandOrder, newTransferRecipient string) error {
	for i := range h {
		err := h[i].AfterDemandOrderFulfilled(ctx, o, newTransferRecipient)
		if err != nil {
			return err
		}
	}
	return nil
}

type BaseEIBCHook struct{}

var _ EIBCHooks = BaseEIBCHook{}

func (b BaseEIBCHook) AfterDemandOrderFulfilled(ctx sdk.Context, o *DemandOrder, newTransferRecipient string) error {
	return nil
}
