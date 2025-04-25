package forward

import (
	"math"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	dackkeeper "github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	types "github.com/dymensionxyz/dymension/v3/x/forward/types"
)

var _ dackkeeper.CompletionHookInstance = rollToHLHook{}

func (k Forward) RollToHLHook() rollToHLHook {
	return rollToHLHook{
		Forward: &k,
	}
}

type rollToHLHook struct {
	*Forward
}

func (h rollToHLHook) ValidateArg(data []byte) error {
	var d types.HookForwardToHL
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
func (h rollToHLHook) Run(ctx sdk.Context, fundsSource sdk.AccAddress, budget sdk.Coin, hookData []byte) error {
	// if fails, the original target got the funds anyway so no need to do anything special (relying on frontend here)
	h.executeWithErrEvent(ctx, func() error {
		var d types.HookForwardToHL
		err := proto.Unmarshal(hookData, &d)
		if err != nil {
			return errorsmod.Wrap(err, "unmarshal")
		}
		return h.forwardToHyperlane(ctx, fundsSource, budget, d)
	})
	return nil
}

var _ dackkeeper.CompletionHookInstance = rollToIBCHook{}

func (k Forward) RollToIBCHook() rollToIBCHook {
	return rollToIBCHook{
		Forward: &k,
	}
}

type rollToIBCHook struct {
	*Forward
}

func (h rollToIBCHook) ValidateArg(data []byte) error {
	var d types.HookForwardToIBC
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
func (h rollToIBCHook) Run(ctx sdk.Context, fundsSource sdk.AccAddress, budget sdk.Coin, hookData []byte) error {
	// if fails, the original target got the funds anyway so no need to do anything special (relying on frontend here)
	h.executeWithErrEvent(ctx, func() error {
		var d types.HookForwardToIBC
		err := proto.Unmarshal(hookData, &d)
		if err != nil {
			return errorsmod.Wrap(err, "unmarshal")
		}
		// funds src is the original ibc transfer recipient, which has now been credited by the eibc fulfiller
		return h.forwardToIBC(ctx, d.Transfer, fundsSource, budget)
	})
	return nil
}

func (k Forward) forwardToIBC(ctx sdk.Context, transfer *ibctransfertypes.MsgTransfer, fundsSrc sdk.AccAddress, maxBudget sdk.Coin) error {

	maxAmt := maxBudget.Amount
	desiredAmt := transfer.Token.Amount
	amt := math.MinInt(maxAmt, desiredAmt)

	m := ibctransfertypes.NewMsgTransfer(
		transfer.SourcePort,
		transfer.SourceChannel,
		maxBudget,
		fundsSrc.String(),
		transfer.Receiver,
		ibcclienttypes.Height{}, // ignore, removed in ibc v2 also
		transfer.TimeoutTimestamp,
		transfer.Memo, // include the original memo, so that we can have more functionality down the road (.e.g actions on rollapp)
	)

	// If this transfer fails asynchronously (timeout or ack) then the funds will get refunded back to the fundSrc by ibc transfer app
	_, err := k.transferK.Transfer(ctx, m)

	return err
}
