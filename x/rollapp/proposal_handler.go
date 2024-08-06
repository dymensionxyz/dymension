package rollapp

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/keeper"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func NewRollappProposalHandler(k *keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.SubmitFraudProposal:
			return HandleSubmitFraudProposal(ctx, k, c)
		case *types.ChangeVMTypeProposal:
			return HandleChangeVMTypeProposal(ctx, k, c)
		default:
			return errorsmod.Wrapf(types.ErrUnknownRequest, "unrecognized rollapp proposal content type: %T", c)
		}
	}
}

func HandleSubmitFraudProposal(ctx sdk.Context, k *keeper.Keeper, p *types.SubmitFraudProposal) error {
	return k.HandleFraud(ctx, p.RollappId, p.IbcClientId, p.FraudelentHeight, p.FraudelentSequencerAddress)
}

func HandleChangeVMTypeProposal(ctx sdk.Context, k *keeper.Keeper, p *types.ChangeVMTypeProposal) error {
	return k.ChangeVMType(ctx, p.RollappId, p.VmType)
}
