package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	commontypes "github.com/dymensionxyz/dymension/v3/x/common/types"
)

type EIBCHooks interface {
	AfterDemandOrderFulfilled(ctx sdk.Context, demandOrder *commontypes.DemandOrder, newTransferRecipient string) error
}

type MultiEIBCHooks []EIBCHooks

var _ EIBCHooks = MultiEIBCHooks{}

func NewMultiEIBCHooks(hooks ...EIBCHooks) MultiEIBCHooks {
	return hooks
}

func (h MultiEIBCHooks) AfterDemandOrderFulfilled(ctx sdk.Context, o *commontypes.DemandOrder, newTransferRecipient string) error {
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

func (b BaseEIBCHook) AfterDemandOrderFulfilled(ctx sdk.Context, o *commontypes.DemandOrder, newTransferRecipient string) error {
	return nil
}
