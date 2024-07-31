package dymns

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	dymnskeeper "github.com/dymensionxyz/dymension/v3/x/dymns/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// NewDymNsProposalHandler creates a governance handler to manage new proposal types.
func NewDymNsProposalHandler(dk dymnskeeper.Keeper) govv1beta1.Handler {
	return func(ctx sdk.Context, content govv1beta1.Content) error {
		switch c := content.(type) {
		case *dymnstypes.MigrateChainIdsProposal:
			return handleMigrateChainIdsProposal(ctx, dk, c)
		case *dymnstypes.UpdateAliasesProposal:
			return handleUpdateAliasesProposal(ctx, dk, c)
		default:
			return errorsmod.Wrapf(errortypes.ErrUnknownRequest, "unrecognized %s proposal content type: %T", dymnstypes.ModuleName, c)
		}
	}
}

// handleMigrateChainIdsProposal handles the proposal to migrate chain id
func handleMigrateChainIdsProposal(
	ctx sdk.Context,
	dk dymnskeeper.Keeper,
	p *dymnstypes.MigrateChainIdsProposal,
) error {
	if err := p.ValidateBasic(); err != nil {
		return err
	}

	err := dk.MigrateChainIds(ctx, p.Replacement)
	if err != nil {
		return err
	}

	return nil
}

// handleUpdateAliasesProposal handles the proposal to update alias for chain-ids
func handleUpdateAliasesProposal(
	ctx sdk.Context,
	dk dymnskeeper.Keeper,
	p *dymnstypes.UpdateAliasesProposal,
) error {
	if err := p.ValidateBasic(); err != nil {
		return err
	}

	err := dk.UpdateAliases(ctx, p.Add, p.Remove)
	if err != nil {
		return err
	}

	return nil
}
