package denommetadata

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/dymensionxyz/dymension/v3/x/denommetadata/keeper"
	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
)

func NewDenomMetadataProposalHandler(k keeper.Keeper) govtypes.Handler {
	return func(ctx sdk.Context, content govtypes.Content) error {
		switch c := content.(type) {
		case *types.CreateDenomMetadataProposal:
			return HandleCreateDenomMetadataProposal(ctx, k, c)
		default:
			return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "unrecognized streamer proposal content type: %T", c)
		}
	}
}

// HandleCreateDenomMetadataProposal is a handler for executing a passed denom metadata proposal
func HandleCreateDenomMetadataProposal(ctx sdk.Context, k keeper.Keeper, p *types.CreateDenomMetadataProposal) error {
	_, err := k.CreateDenomMetadata(ctx, p.TokenMetadata)
	if err != nil {
		return err
	}
	return nil
}
