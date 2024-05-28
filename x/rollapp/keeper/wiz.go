package keeper

import (
	"encoding/json"
	"errors"
	"fmt"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/dymensionxyz/dymension/v3/utils"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func ParseSpecialTransferDenom(memo string) (*banktypes.Metadata, error) {
	type T struct {
		// If the packet originates from the chain itself, and not a user, this will be true
		DoesNotOriginateFromUser bool               `json:"does_not_originate_from_user"`
		DenomMetadata            banktypes.Metadata `json:"denom_metadata"`
	}
	var t T
	err := json.Unmarshal([]byte(memo), &t)
	if err != nil {
		return nil, sdkerrors.ErrJSONUnmarshal
	}
	if !t.DoesNotOriginateFromUser {
		return nil, sdkerrors.ErrUnauthorized
	}
	return &t.DenomMetadata, nil
}

func (k Keeper) MarkGenesisAsHappened(ctx sdktypes.Context, args types.TriggerGenesisArgs) error {
	// Get the rollapp
	rollapp, found := k.GetRollapp(ctx, args.RollappID)
	if !found {
		panic("expected to find rollapp")
	}

	// Validate it hasn't been triggered yet
	if rollapp.GenesisState.GenesisEventHappened {
		panic(errors.New("genesis event already happened - it shouldn't have"))
	}

	rollapp.GenesisState.GenesisEventHappened = true
	k.SetRollapp(ctx, rollapp)

	return nil
}

func (k Keeper) RegisterOneDenomMetadata(ctx sdktypes.Context, md *types.TokenMetadata, rollappID, channelID string) error {
	denomTrace := utils.GetForeignDenomTrace(channelID, md.Base)
	traceHash := denomTrace.Hash()
	// if the denom trace does not exist, add it
	if !k.transferKeeper.HasDenomTrace(ctx, traceHash) {
		k.transferKeeper.SetDenomTrace(ctx, denomTrace)
	}

	ibcBaseDenom := denomTrace.IBCDenom()

	// create a new token denom metadata where it's base = ibcDenom,
	// and the rest of the fields are taken from Metadata
	metadata := banktypes.Metadata{
		Description: "auto-generated metadata for " + ibcBaseDenom + " from rollapp " + rollappID,
		Base:        ibcBaseDenom,
		DenomUnits:  make([]*banktypes.DenomUnit, len(md.DenomUnits)),
		Display:     md.Display,
		Name:        md.Name,
		Symbol:      md.Symbol,
		URI:         md.URI,
		URIHash:     md.URIHash,
	}
	// Copy DenomUnits slice
	for j, du := range md.DenomUnits {
		newDu := banktypes.DenomUnit{
			Aliases:  du.Aliases,
			Denom:    du.Denom,
			Exponent: du.Exponent,
		}
		// base denom_unit should be the same as baseDenom
		if newDu.Exponent == 0 {
			newDu.Denom = ibcBaseDenom
			newDu.Aliases = append(newDu.Aliases, du.Denom)
		}
		metadata.DenomUnits[j] = &newDu
	}

	// validate metadata
	if validity := metadata.Validate(); validity != nil {
		return fmt.Errorf("invalid denom metadata on genesis event: %w", validity)
	}

	// save the new token denom metadata
	if err := k.denommetadataKeeper.CreateDenomMetadata(ctx, metadata); err != nil {
		return fmt.Errorf("create denom metadata: %w", err)
	}

	k.Logger(ctx).Info("registered denom metadata for IBC token", "rollappID", rollappID, "denom", ibcBaseDenom)
	return nil
}
