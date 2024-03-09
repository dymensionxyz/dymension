package denommetadata

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/dymensionxyz/dymension/v3/x/denommetadata/types"
)

func NewDenomMetadataProposalHandler(k types.BankKeeper) govtypes.Handler {
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
func HandleCreateDenomMetadataProposal(ctx sdk.Context, k types.BankKeeper, p *types.CreateDenomMetadataProposal) error {

	found := k.HasDenomMetaData(ctx, p.TokenMetadata.Base)
	if found {
		return types.ErrDenomAlreadyExists
	}

	k.SetDenomMetaData(ctx, p.TokenMetadata)
	return nil
}

// HandleUpdateDenomMetadataProposal is a handler for executing a passed denom metadata update proposal
func HandleUpdateDenomMetadataProposal(ctx sdk.Context, k types.BankKeeper, p *types.UpdateDenomMetadataProposal) error {

	found := k.HasDenomMetaData(ctx, p.TokenMetadata.Base)
	if !found {
		return types.ErrDenomDoesNotExist
	}

	k.SetDenomMetaData(ctx, p.TokenMetadata)
	return nil
}
