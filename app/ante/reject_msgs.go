package ante

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/authz"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/group"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
)

// RejectMessagesDecorator prevents invalid msg types from being executed
type RejectMessagesDecorator struct {
	disabledMsgTypeURLs map[string]struct{}
	cdc                 codec.Codec
	groupKeeper         *groupkeeper.Keeper
}

var _ sdk.AnteDecorator = RejectMessagesDecorator{}

// NewRejectMessagesDecorator creates a decorator to block vesting messages from reaching the mempool
func NewRejectMessagesDecorator(cdc codec.Codec, groupKeeper *groupkeeper.Keeper, disabledMsgTypeURLs ...string) RejectMessagesDecorator {
	disabledMsgsMap := make(map[string]struct{})
	for _, url := range disabledMsgTypeURLs {
		disabledMsgsMap[url] = struct{}{}
	}

	return RejectMessagesDecorator{
		disabledMsgTypeURLs: disabledMsgsMap,
		cdc:                 cdc,
		groupKeeper:         groupKeeper,
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
	if nestedMsgs >= maxNestedMsgs {
		return fmt.Errorf("found more nested msgs than permitted. Limit is : %d", maxNestedMsgs)
	}

	for _, msg := range msgs {
		typeURL := sdk.MsgTypeURL(msg)
		if rmd.isDisabledMsg(typeURL) {
			return fmt.Errorf("found disabled msg type: %s", typeURL)
		}

		if _, ok := msg.(*evmtypes.MsgEthereumTx); ok {
			return errorsmod.Wrapf(
				sdkerrors.ErrInvalidType,
				"MsgEthereumTx needs to be contained within a tx with 'ExtensionOptionsEthereumTx' option",
			)
		}

		switch concreteMsg := msg.(type) {
		case *authz.MsgExec:
			innerMsgs, err := concreteMsg.GetMessages()
			if err != nil {
				return err
			}
			nestedMsgs++
			if err := rmd.checkMsgs(ctx, innerMsgs, nestedMsgs); err != nil {
				return err
			}
		case *authz.MsgGrant:
			authorization, err := concreteMsg.GetAuthorization()
			if err != nil {
				return err
			}
			nestedMsgs++
			url := authorization.MsgTypeURL()
			if rmd.isDisabledMsg(url) {
				return fmt.Errorf("granting disabled msg type: %s is not allowed", url)
			}
		case *govtypesv1.MsgSubmitProposal:
			nestedMsgs++
			if err := rmd.checkMessages(ctx, concreteMsg.Messages, nestedMsgs); err != nil {
				return err
			}
		case *group.MsgExec:
			proposalRes, err := rmd.groupKeeper.Proposal(
				sdk.WrapSDKContext(ctx),
				&group.QueryProposalRequest{ProposalId: concreteMsg.ProposalId},
			)
			if err != nil {
				return err
			}
			if proposalRes.Proposal == nil {
				return errorsmod.Wrapf(sdkerrors.ErrInvalidRequest, "proposal with ID %d not found", concreteMsg.ProposalId)
			}
			nestedMsgs++
			if err := rmd.checkMessages(ctx, proposalRes.Proposal.Messages, nestedMsgs); err != nil {
				return err
			}
		case *group.MsgSubmitProposal:
			nestedMsgs++
			if err := rmd.checkMessages(ctx, concreteMsg.Messages, nestedMsgs); err != nil {
				return err
			}
		default:
		}
	}
	return nil
}

func (rmd RejectMessagesDecorator) checkMessages(ctx sdk.Context, anyMsgs []*codectypes.Any, nestedMsgs int) error {
	for _, anyMsg := range anyMsgs {
		var innerMsg sdk.Msg
		if err := rmd.cdc.UnpackAny(anyMsg, &innerMsg); err != nil {
			return err
		}
		if err := rmd.checkMsgs(ctx, []sdk.Msg{innerMsg}, nestedMsgs); err != nil {
			return err
		}
	}
	return nil
}

func (rmd RejectMessagesDecorator) isDisabledMsg(msgTypeURL string) bool {
	_, ok := rmd.disabledMsgTypeURLs[msgTypeURL]
	return ok
}
