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

// at the time of calling, funds have either been sent from the eibc LP to the ibc transfer recipient, or minted/unescrowed from
// the ibc transfer app to the ibc transfer recipient
func (h rollappToHubCompletion) Run(ctx sdk.Context, fundsSource sdk.AccAddress, budget sdk.Coin, hookData []byte) error {
	// if fails, the original target got the funds anyway so no need to do anything special (relying on frontend here)
	h.forwardWithEvent(ctx, func() error {
		var d types.HookEIBCtoHL
		err := proto.Unmarshal(hookData, &d)
		if err != nil {
			return errorsmod.Wrap(err, "unmarshal")
		}
		return h.forwardToHyperlane(ctx, fundsSource, budget, d)
	})
	return nil
}

func (k Keeper) forwardToIBC(ctx sdk.Context, transfer *ibctransfertypes.MsgTransfer, fundsSrc sdk.AccAddress, maxBudget sdk.Coin) error {

	m := ibctransfertypes.NewMsgTransfer(
		transfer.SourcePort,
		transfer.SourceChannel,
		maxBudget,
		fundsSrc.String(),
		transfer.Receiver,
		ibcclienttypes.Height{}, // ignore, removed in ibc v2 also
		transfer.TimeoutTimestamp,
		transfer.Memo,
	)

	// If this transfer fails asynchronously (timeout or ack) then the funds will get refunded back to the fundSrc by ibc transfer app
	_, err := k.transferK.Transfer(ctx, m)

	return err
}
