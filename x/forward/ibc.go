package forward

import (
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	ibctransfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	dackkeeper "github.com/dymensionxyz/dymension/v3/x/delayedack/keeper"
	types "github.com/dymensionxyz/dymension/v3/x/forward/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
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
	h.executeAtomicWithErrEvent(ctx, func(c sdk.Context) (bool, error) {
		var d types.HookForwardToHL
		err := proto.Unmarshal(hookData, &d)
		if err != nil {
			return true, errorsmod.Wrap(err, "unmarshal")
		}
		return true, h.forwardToHyperlane(c, fundsSource, budget, d)
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
	h.executeAtomicWithErrEvent(ctx, func(c sdk.Context) (bool, error) {
		var d types.HookForwardToIBC
		err := proto.Unmarshal(hookData, &d)
		if err != nil {
			return true, errorsmod.Wrap(err, "unmarshal")
		}
		// funds src is the original ibc transfer recipient, which has now been credited by the eibc fulfiller
		return true, h.forwardToIBC(c, d.Transfer, fundsSource, budget, d.MinAmount)
	})
	return nil
}

func (k Forward) forwardToIBC(ctx sdk.Context, transfer *ibctransfertypes.MsgTransfer, fundsSrc sdk.AccAddress, maxBudget sdk.Coin, minAmount math.Int) error {
	if !minAmount.IsNil() && minAmount.IsPositive() && maxBudget.Amount.LT(minAmount) {
		return gerrc.ErrInvalidArgument.Wrapf("forwardable budget %s below min_amount %s", maxBudget.Amount, minAmount)
	}

	nowNs := uint64(ctx.BlockTime().UnixNano()) //nolint:gosec // block time is never negative
	timeoutTimestamp := transfer.TimeoutTimestamp
	if timeoutTimestamp == 0 || timeoutTimestamp <= nowNs+types.MinForwardIBCTimeout {
		// Give the final hop a fresh deadline when the composer's absolute timeout is no longer usable.
		timeoutTimestamp = nowNs + types.DefaultForwardIBCTimeout
	}

	m := ibctransfertypes.NewMsgTransfer(
		transfer.SourcePort,
		transfer.SourceChannel,
		maxBudget,
		fundsSrc.String(),
		transfer.Receiver,
		ibcclienttypes.Height{}, // ignore, removed in ibc v2 also
		timeoutTimestamp,
		transfer.Memo, // include the original memo, so that we can have more functionality down the road (.e.g actions on rollapp)
	)

	// If this transfer fails asynchronously (timeout or ack) then the funds will get refunded back to the fundSrc by ibc transfer app
	_, err := k.transferK.Transfer(ctx, m)

	return err
}
