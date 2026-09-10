package ante

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	circuitante "cosmossdk.io/x/circuit/ante"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/group"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

var _ sdk.AnteDecorator = RejectMessagesDecorator{}

// RejectMessagesDecorator prevents invalid msg types from being executed
type RejectMessagesDecorator struct {
	// message is rejected if any Predicate returns true
	predicates []Predicate
}

// Predicate should return true if message is not allowed.
// A non nil error means the policy could not be evaluated at all, which is distinct from
// the message being disallowed, and must not be reported as the latter.
type Predicate = func(ctx sdk.Context, typeURL string, depth int) (bool, error)

// Blocks any message with depth of depthMax OR MORE
// Depth 0 is top level message
// Depth 1 or more is wrapped in something
func BlockTypeUrls(depthMax int, typeUrls ...string) Predicate {
	block := make(map[string]struct{})
	for _, url := range typeUrls {
		block[url] = struct{}{}
	}
	return func(_ sdk.Context, url string, depth int) (bool, error) {
		_, ok := block[url]
		return ok && depthMax <= depth, nil
	}
}

// BlockTrippedByCircuitBreaker blocks, at any depth, messages whose type is currently tripped
// in the circuit breaker. The SDK's own CircuitBreakerDecorator deliberately only inspects
// top level messages, so a tripped type wrapped in authz/gov/group would otherwise be admitted
// to the mempool and only fail later in baseapp's msg router.
func BlockTrippedByCircuitBreaker(ck circuitante.CircuitBreaker) Predicate {
	return func(ctx sdk.Context, url string, _ int) (bool, error) {
		allowed, err := ck.IsAllowed(ctx, url)
		if err != nil {
			// fail closed, but the caller surfaces err rather than claiming a deliberate trip
			return true, err
		}
		return !allowed, nil
	}
}

func NewRejectMessagesDecorator() *RejectMessagesDecorator {
	return &RejectMessagesDecorator{
		predicates: []Predicate{},
	}
}

func (rmd *RejectMessagesDecorator) WithPredicate(p Predicate) *RejectMessagesDecorator {
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
		return ctx, errors.Join(sdkerrors.ErrUnauthorized, err)
	}
	return next(ctx, tx, simulate)
}

// depth=0 means top level message
func (rmd RejectMessagesDecorator) checkMsgs(ctx sdk.Context, msgs []sdk.Msg, depth int) error {
	for _, msg := range msgs {
		if err := rmd.checkMsg(ctx, msg, depth); err != nil {
			return err
		}
	}
	return nil
}

// rejection labels the reason, e.g. 'disabled' for a msg and 'disabled grant' for what a grant authorizes
func (rmd RejectMessagesDecorator) checkPredicates(ctx sdk.Context, typeURL string, depth int, rejection string) error {
	for _, pred := range rmd.predicates {
		blocked, err := pred(ctx, typeURL, depth)
		if err != nil {
			return errorsmod.Wrapf(err, "check msg: %s", typeURL)
		}
		if blocked {
			return gerrc.ErrInvalidArgument.Wrapf("%s: %s", rejection, typeURL)
		}
	}
	return nil
}

// depth=0 means top level message
func (rmd RejectMessagesDecorator) checkMsg(ctx sdk.Context, msg sdk.Msg, depth int) error {
	if depth >= maxInnerDepth {
		return fmt.Errorf("found more nested msgs than permitted. limit is : %d", maxInnerDepth)
	}

	if _, ok := msg.(*evmtypes.MsgEthereumTx); ok {
		return errorsmod.Wrapf(
			sdkerrors.ErrInvalidType,
			"MsgEthereumTx needs to be contained within a tx with 'ExtensionOptionsEthereumTx' option",
		)
	}

	if err := rmd.checkPredicates(ctx, sdk.MsgTypeURL(msg), depth, "disabled"); err != nil {
		return err
	}

	var err error
	var inner []sdk.Msg

	switch m := msg.(type) {
	case *authz.MsgExec:
		inner, err = m.GetMessages()
	case *govtypesv1.MsgSubmitProposal:
		inner, err = m.GetMsgs()
	case *group.MsgSubmitProposal:
		inner, err = m.GetMsgs()
	case *authz.MsgGrant:
		authorization, err := m.GetAuthorization()
		if err != nil {
			return err
		}
		if err := rmd.checkPredicates(ctx, authorization.MsgTypeURL(), depth, "disabled grant"); err != nil {
			return err
		}
	default:
	}

	if err != nil {
		return err
	}

	return rmd.checkMsgs(ctx, inner, depth+1)
}
