package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	types "github.com/dymensionxyz/dymension/v3/x/forward/types"
	transfer "github.com/dymensionxyz/dymension/v3/x/transfer"
)

var _ transfer.CompletionHookInstance = rollappToHubCompletion{}

func (k Keeper) Hook() rollappToHubCompletion {
	return rollappToHubCompletion{
		Keeper: &k,
	}
}

type rollappToHubCompletion struct {
	*Keeper
}

func (h rollappToHubCompletion) ValidateData(data []byte) error {
	var d types.HookEIBCtoHL
	err := proto.Unmarshal(data, &d)
	if err != nil {
		return errorsmod.Wrap(err, "unmarshal")
	}
	if err := d.ValidateBasic(); err != nil {
		return errorsmod.Wrap(err, "validate")
	}
	return nil
}

// TODO: rename method
// at this point funds have not been sent from the fulfiller/eibc LP/funds provider to the recipient (or anywhere else)
func (h rollappToHubCompletion) Run(ctx sdk.Context, fundsSource sdk.AccAddress, originalTransferRecipient sdk.AccAddress, budget sdk.Coin, hookData []byte) error {

	h.refundOnError(ctx, func() error {
		var d types.HookEIBCtoHL
		err := proto.Unmarshal(hookData, &d)
		if err != nil {
			return errorsmod.Wrap(err, "unmarshal")
		}

		return h.forwardToHyperlane(ctx, fundsSource, budget, d)
	}, fundsSource, originalTransferRecipient, budget)
	return nil
}

func (k Keeper) forwardToIBC(ctx sdk.Context, transfer *ibctransfertypes.MsgTransfer, fundsSrc sdk.AccAddress, maxBudget sdk.Coin, memo []byte) error {

	m := ibctransfertypes.NewMsgTransfer(
		transfer.SourcePort,
		transfer.SourceChannel,
		maxBudget,
		fundsSrc.String(),
		transfer.Receiver,
		ibcclienttypes.Height{}, // ignore, removed in ibc v2 also
		transfer.TimeoutTimestamp,
		string(memo),
	)

	// If this transfer fails asynchronously (timeout or ack) then the funds will get refunded back to the sender by ibc transfer app
	_, err := k.transferK.Transfer(ctx, m)

	return err
}
