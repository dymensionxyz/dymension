package denommetadata

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"

	"github.com/dymensionxyz/dymension/v3/x/denommetadata/keeper"
	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
)

func NewDenomMetadataProposalHandler(k *keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.CreateDenomMetadataProposal:
			return HandleCreateDenomMetadataProposal(ctx, k, c)
		case *types.UpdateDenomMetadataProposal:
			return HandleUpdateDenomMetadataProposal(ctx, k, c)
		default:
			return errorsmod.WithType(gerrc.ErrInvalidArgument, c)
		}
	}
}

// HandleCreateDenomMetadataProposal is a handler for executing a passed denom metadata creation proposal
func HandleCreateDenomMetadataProposal(ctx sdk.Context, k *keeper.Keeper, p *types.CreateDenomMetadataProposal) error {
	if err := p.ValidateBasic(); err != nil {
		return err
	}

	for _, metadata := range p.TokenMetadata {
		err := k.CreateDenomMetadata(ctx, metadata)
		if err != nil {
			return err
		}
	}
	return nil
}

// HandleUpdateDenomMetadataProposal is a handler for executing a passed denom metadata update proposal
func HandleUpdateDenomMetadataProposal(ctx sdk.Context, k *keeper.Keeper, p *types.UpdateDenomMetadataProposal) error {
	if err := p.ValidateBasic(); err != nil {
		return err
	}

	for _, metadata := range p.TokenMetadata {
		err := k.UpdateDenomMetadata(ctx, metadata)
		if err != nil {
			return err
		}
	}
	return nil
}
