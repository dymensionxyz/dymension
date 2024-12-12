package sequencer

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/dymensionxyz/dymension/v3/x/sequencer/keeper"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

func NewSequencerProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.PunishSequencerProposal:
			return HandlePunishSequencerProposal(ctx, k, c)
		default:
			return errorsmod.Wrapf(types.ErrUnknownRequest, "unrecognized sequencer proposal content type: %T", c)
		}
	}
}

func HandlePunishSequencerProposal(ctx sdk.Context, k keeper.Keeper, p *types.PunishSequencerProposal) error {
	err := k.PunishSequencer(ctx, p.PunishSequencerAddress, p.MustRewardee())
	if err != nil {
		k.Logger(ctx).Error("failed to punish sequencer", "error", err)
		return err
	}
	return nil
}
