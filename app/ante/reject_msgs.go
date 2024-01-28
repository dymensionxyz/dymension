package ante

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
)

// RejectMessagesDecorator prevents invalid msg types from being executed
type RejectMessagesDecorator struct {
	disabledMsgTypeURLs []string
}

var _ sdk.AnteDecorator = RejectMessagesDecorator{}

// NewRejectMessagesDecorator creates a decorator to block vesting messages from reaching the mempool
func NewRejectMessagesDecorator() RejectMessagesDecorator {
	return RejectMessagesDecorator{
		disabledMsgTypeURLs: []string{
			sdk.MsgTypeURL(&vesting.MsgCreateVestingAccount{}),
			sdk.MsgTypeURL(&vesting.MsgCreatePermanentLockedAccount{}),
			sdk.MsgTypeURL(&vesting.MsgCreatePeriodicVestingAccount{}),
		},
	}
}

// AnteHandle rejects messages that requires ethereum-specific authentication.
// For example `MsgEthereumTx` requires fee to be deducted in the antehandler in
// order to perform the refund.
func (rmd RejectMessagesDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	for _, msg := range tx.GetMsgs() {
		if _, ok := msg.(*evmtypes.MsgEthereumTx); ok {
			return ctx, errorsmod.Wrapf(
				sdkerrors.ErrInvalidType,
				"MsgEthereumTx needs to be contained within a tx with 'ExtensionOptionsEthereumTx' option",
			)
		}

		typeURL := sdk.MsgTypeURL(msg)
		for _, disabledTypeURL := range rmd.disabledMsgTypeURLs {
			if typeURL == disabledTypeURL {
				return ctx, errorsmod.Wrapf(
					sdkerrors.ErrUnauthorized,
					"MsgTypeURL %s not supported",
					typeURL,
				)
			}
		}
	}
	return next(ctx, tx, simulate)
}
