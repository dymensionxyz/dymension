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

const (
	HookNameForward = "forward"
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
	err := proto.Unmarshal(order.FulfillHook.HookData, &d)
	if err != nil {
		return errorsmod.Wrap(err, "unmarshal forward hook")
	}
	err = k.escrowFromUser(ctx, fundsSource, order.Price)
	if err != nil {
		// should never happen
		err = errorsmod.Wrap(err, "escrow from user")
		k.Logger(ctx).Error("doForwardHook", "error", err)
		return err
	}
	k.forwardToHyperlane(ctx, order, d)
	return nil
}

func (k Keeper) forwardToIBC(ctx sdk.Context, transfer *ibctransfertypes.MsgTransfer, r types.Recovery) {
	k.refundOnError(ctx, func() error {
		return k.forwardToIBCInner(ctx, transfer)
	}, r, sdk.NewCoins(transfer.Token))
}

func (k Keeper) forwardToIBCInner(ctx sdk.Context, transfer *ibctransfertypes.MsgTransfer) error {

	var (
		token            = transfer.Token
		sender           = transfer.Sender
		recipient        = transfer.Receiver
		timeoutTimestamp = transfer.TimeoutTimestamp
		memo             string
	)
	ibctransfertypes.NewMsgTransfer(
		"transfer",
		"channel-0",
		token,
		sender,
		recipient,
		ibcclienttypes.Height{}, // ignore, removed in ibc v2 also
		timeoutTimestamp,
		memo,
	)
	res, err := k.transferKeeper.Transfer(ctx, transfer)
	if err != nil {
		return errorsmod.Wrap(err, "transfer")
	}
	_ = res
	// TODO:
	return nil
}
