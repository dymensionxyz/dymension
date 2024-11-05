package ante

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/group"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

// RejectMessagesDecorator prevents invalid msg types from being executed
type RejectMessagesDecorator struct {
	disabledMsgTypeURLs map[string]struct{}
}

var _ sdk.AnteDecorator = RejectMessagesDecorator{}

// NewRejectMessagesDecorator creates a decorator to block provided messages from reaching the mempool
func NewRejectMessagesDecorator(disabledMsgTypeURLs ...string) RejectMessagesDecorator {
	disabledMsgsMap := make(map[string]struct{})
	for _, url := range disabledMsgTypeURLs {
		disabledMsgsMap[url] = struct{}{}
	}

	return RejectMessagesDecorator{
		disabledMsgTypeURLs: disabledMsgsMap,
	}
}

// AnteHandle recursively rejects messages such as those that requires ethereum-specific authentication.
// For example `MsgEthereumTx` requires fee to be deducted in the ante handler in
// order to perform the refund.
func (rmd RejectMessagesDecorator) AnteHandle(
	ctx sdk.Context,
	tx sdk.Tx,
	simulate bool,
	next sdk.AnteHandler,
) (sdk.Context, error) {
	if err := rmd.checkMsgs(ctx, tx.GetMsgs(), 0); err != nil {
		return ctx, errorsmod.Wrapf(sdkerrors.ErrUnauthorized, err.Error())
	}
	return next(ctx, tx, simulate)
}

const maxNestedMsgs = 6

func (rmd RejectMessagesDecorator) checkMsgs(ctx sdk.Context, msgs []sdk.Msg, nestedMsgs int) error {
	for _, msg := range msgs {
		if err := rmd.checkMsg(ctx, msg, nestedMsgs); err != nil {
			return err
		}
	}
	return nil
}

func (rmd RejectMessagesDecorator) checkMsg(ctx sdk.Context, msg sdk.Msg, nestedMsgs int) error {
	typeURL := sdk.MsgTypeURL(msg)
	if _, ok := rmd.disabledMsgTypeURLs[typeURL]; ok {
		return fmt.Errorf("found disabled msg type: %s", typeURL)
	}

	if _, ok := msg.(*evmtypes.MsgEthereumTx); ok {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidType,
			"MsgEthereumTx needs to be contained within a tx with 'ExtensionOptionsEthereumTx' option",
		)
	}

	if nestedMsgs >= maxNestedMsgs {
		return fmt.Errorf("found more nested msgs than permitted. Limit is : %d", maxNestedMsgs)
	}

	innerMsgs, err := extractMsgs(msg)
	if err != nil {
		return err
	}
	switch concreteMsg := msg.(type) {
	case *authz.MsgExec:
		nestedMsgs++
		if err := rmd.checkMsgs(ctx, innerMsgs, nestedMsgs); err != nil {
			return err
		}
	case *authz.MsgGrant:
		authorization, err := concreteMsg.GetAuthorization()
		if err != nil {
			return err
		}
		url := authorization.MsgTypeURL()
		if _, ok := rmd.disabledMsgTypeURLs[url]; ok {
			return fmt.Errorf("granting disabled msg type: %s is not allowed", url)
		}
	case *govtypesv1.MsgSubmitProposal:
		nestedMsgs++
		if err := rmd.checkMsgs(ctx, innerMsgs, nestedMsgs); err != nil {
			return err
		}
	case *group.MsgSubmitProposal:
		nestedMsgs++
		if err := rmd.checkMsgs(ctx, innerMsgs, nestedMsgs); err != nil {
			return err
		}
	default:
	}

	return nil
}

func extractMsgs(msg any) ([]sdk.Msg, error) {
	if msgWithMsgs, ok := msg.(interface{ GetMsgs() ([]sdk.Msg, error) }); ok {
		return msgWithMsgs.GetMsgs()
	}
	if msgWithMessages, ok := msg.(interface{ GetMessages() ([]sdk.Msg, error) }); ok {
		return msgWithMessages.GetMessages()
	}
	return nil, nil
}
