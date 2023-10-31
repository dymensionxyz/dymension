package lockdrop

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	"github.com/dymensionxyz/dymension/x/lockdrop/keeper"
	"github.com/dymensionxyz/dymension/x/lockdrop/types"
)

// NewPoolIncentivesProposalHandler is a handler for governance proposals on new pool incentives.
func NewPoolIncentivesProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.UpdateLockdropProposal:
			return handleUpdateLockdropProposal(ctx, k, c)
		case *types.ReplaceLockdropProposal:
			return handleReplaceLockdropProposal(ctx, k, c)

		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized pool incentives proposal content type: %T", c)
		}
	}
}

// handleReplacePoolIncentivesProposal is a handler for replacing pool incentives governance proposals
func handleReplaceLockdropProposal(ctx sdk.Context, k keeper.Keeper, p *types.ReplaceLockdropProposal) error {
	return k.HandleReplaceLockdropProposal(ctx, p)
}

// handleUpdateLockdropProposal is a handler for updating pool incentives governance proposals
func handleUpdateLockdropProposal(ctx sdk.Context, k keeper.Keeper, p *types.UpdateLockdropProposal) error {
	return k.HandleUpdateLockdropProposal(ctx, p)
}
