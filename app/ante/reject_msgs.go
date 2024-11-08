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
	// message is rejected if any predicate returns true
	predicates []predicate
}

// predicate should return true if message is not allowed
type predicate = func(typeURL string, depth int) bool

func blockTypeUrls(typeUrls ...string) predicate {
	block := make(map[string]struct{})
	for _, url := range typeUrls {
		block[url] = struct{}{}
	}
	return func(url string, depth int) bool {
		_, ok := block[url]
		return ok
	}
}

var _ sdk.AnteDecorator = RejectMessagesDecorator{}

func NewRejectMessagesDecorator() *RejectMessagesDecorator {
	return &RejectMessagesDecorator{
		predicates: []predicate{},
	}
}

func (rmd *RejectMessagesDecorator) withPredicate(p predicate) *RejectMessagesDecorator {
	rmd.predicates = append(rmd.predicates, p)
	return rmd
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

const maxDepth = 6

// depth=0 means top level message
func (rmd RejectMessagesDecorator) checkMsgs(ctx sdk.Context, msgs []sdk.Msg, depth int) error {
	for _, msg := range msgs {
		if err := rmd.checkMsg(ctx, msg, depth); err != nil {
			return err
		}
	}
	return nil
}

// depth=0 means top level message
func (rmd RejectMessagesDecorator) checkMsg(ctx sdk.Context, msg sdk.Msg, depth int) error {
	if depth >= maxDepth {
		return fmt.Errorf("found more nested msgs than permitted. Limit is : %d", maxDepth)
	}

	if _, ok := msg.(*evmtypes.MsgEthereumTx); ok {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidType,
			"MsgEthereumTx needs to be contained within a tx with 'ExtensionOptionsEthereumTx' option",
		)
	}

	typeURL := sdk.MsgTypeURL(msg)
	for _, pred := range rmd.predicates {
		if pred(typeURL, depth) {
			return fmt.Errorf("found disabled msg type: %s", typeURL)
		}
	}

	depth++

	var err error
	var inner []sdk.Msg

	switch m := msg.(type) {
	case *authz.MsgExec:
		inner, err = m.GetMessages()
	case *govtypesv1.MsgSubmitProposal:
		inner, err = m.GetMsgs()
	case *group.MsgSubmitProposal:
		inner, err = m.GetMsgs()
	default:
	}

	if err != nil {
		return err
	}

	return rmd.checkMsgs(ctx, inner, depth+1)
}
