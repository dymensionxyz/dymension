package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	eibckeeper "github.com/dymensionxyz/dymension/v3/x/eibc/keeper"
	eibctypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
	types "github.com/dymensionxyz/dymension/v3/x/forward/types"
)

var _ eibckeeper.FulfillHook = eIBCHook{}

func (k Keeper) Hook() eIBCHook {
	return eIBCHook{
		Keeper: &k,
	}
}

type eIBCHook struct {
	*Keeper
}

func (h eIBCHook) ValidateData(data []byte) error {
	return validEIBCForward(data)
}

func validEIBCForward(data []byte) error {
	var d types.HookEIBCtoHL
	err := proto.Unmarshal(data, &d)
	if err != nil {
		return errorsmod.Wrap(err, "unmarshal forward hook")
	}
	if err := d.ValidateBasic(); err != nil {
		return errorsmod.Wrap(err, "validate forward hook")
	}
	return nil

}

func (h eIBCHook) Run(ctx sdk.Context, order *eibctypes.DemandOrder, fundsSource sdk.AccAddress,
	newTransferRecipient sdk.AccAddress,
	fulfiller sdk.AccAddress, hookData []byte) error {
	return h.onEIBC(ctx, order, fundsSource, hookData)
}

// at this point funds have not been sent from the fulfiller/eibc LP/funds provider to the recipient (or anywhere else)
func (k Keeper) onEIBC(ctx sdk.Context, order *eibctypes.DemandOrder, fundsSource sdk.AccAddress, data []byte) error {
	var d types.HookEIBCtoHL
	err := proto.Unmarshal(data, &d)
	if err != nil {
		return errorsmod.Wrap(err, "unmarshal forward hook")
	}
	err = k.escrowFromAccount(ctx, fundsSource, order.Price)
	if err != nil {
		// should never happen
		err = errorsmod.Wrap(err, "escrow from user")
		k.Logger(ctx).Error("doForwardHook", "error", err)
		return err
	}
	k.forwardToHyperlane(ctx, order, d)
	return nil
}

func (k Keeper) forwardToIBC(ctx sdk.Context, transfer *ibctransfertypes.MsgTransfer, r types.Recovery, maxBudget sdk.Coin, memo []byte) {
	k.refundOnError(ctx, func() error {
		return k.forwardToIBCInner(ctx, transfer, maxBudget, memo)
	}, r, sdk.NewCoins(transfer.Token))
}

func (k Keeper) forwardToIBCInner(ctx sdk.Context, transfer *ibctransfertypes.MsgTransfer, maxBudget sdk.Coin, memo []byte) error {

	ibctransfertypes.NewMsgTransfer(
		transfer.SourcePort,
		transfer.SourceChannel,
		maxBudget,
		k.getModuleAddr(ctx).String(),
		transfer.Receiver,
		ibcclienttypes.Height{}, // ignore, removed in ibc v2 also
		transfer.TimeoutTimestamp,
		string(memo), // TODO: conversion ok?
	)

	_, err := k.transferK.Transfer(ctx, transfer) // TODO: response?
	return err
}
