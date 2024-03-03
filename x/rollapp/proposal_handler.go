package rollapp

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"

	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

func NewRollappProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.SubmitFraudProposal:
			return HandleSubmitFraudProposal(ctx, k, c)
		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized rollapp proposal content type: %T", c)
		}
	}
}

func HandleSubmitFraudProposal(ctx sdk.Context, k keeper.Keeper, p *types.SubmitFraudProposal) error {
	err := k.HandleFraud(ctx, p.RollappId, p.FraudelentHeight, p.FraudelentSequencerAddress)
	if err != nil {
		return err
	}
	return nil
}
