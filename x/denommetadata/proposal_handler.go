package denommetadata

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
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
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized denommetadata proposal content type: %T", c)
		}
	}
}

// HandleCreateDenomMetadataProposal is a handler for executing a passed denom metadata creation proposal
func HandleCreateDenomMetadataProposal(ctx sdk.Context, k *keeper.Keeper, p *types.CreateDenomMetadataProposal) error {

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

	for _, metadata := range p.TokenMetadata {
		err := k.UpdateDenomMetadata(ctx, metadata)
		if err != nil {
			return err
		}
	}
	return nil
}
