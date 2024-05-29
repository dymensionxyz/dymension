package keeper

import (
	"fmt"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/dymensionxyz/dymension/v3/utils"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k Keeper) MarkGenesisAsHappened(ctx sdktypes.Context, args types.TriggerGenesisArgs) error {
	// Get the rollapp
	rollapp, found := k.GetRollapp(ctx, args.RollappID)
	if !found {
		panic("expected to find rollapp")
	}

	// Validate it hasn't been triggered yet
	if rollapp.GenesisState.GenesisEventHappened {
		k.Logger().Error("genesis event already happened")
		// panic(errors.New("genesis event already happened - it shouldn't have"))
	}

	rollapp.GenesisState.GenesisEventHappened = true
	k.SetRollapp(ctx, rollapp)

	return nil
}

func (k Keeper) RegisterOneDenomMetadata(ctx sdktypes.Context, m banktypes.Metadata, rollappID, channelID string) error {
	denomTrace := utils.GetForeignDenomTrace(channelID, m.Base)
	traceHash := denomTrace.Hash()
	// if the denom trace does not exist, add it
	if !k.transferKeeper.HasDenomTrace(ctx, traceHash) {
		// TODO: why do we have this check? why not just set?
		k.transferKeeper.SetDenomTrace(ctx, denomTrace)
	}

	ibcDenom := denomTrace.IBCDenom()

	/*
		Change the base to the ibc denom, and add an alias to the original
	*/
	m.Description = fmt.Sprintf("auto-generated ibc denom for rollapp: base: %s: rollapp: %s", ibcDenom, rollappID)
	m.Base = ibcDenom
	for i, u := range m.DenomUnits {
		if u.Exponent == 0 {
			m.DenomUnits[i].Aliases = append(m.DenomUnits[i].Aliases, u.Denom)
			m.DenomUnits[i].Denom = ibcDenom

		}
	}

	// validate metadata
	if validity := m.Validate(); validity != nil {
		return fmt.Errorf("invalid denom metadata on genesis event: %w", validity)
	}

	// save the new token denom metadata
	if err := k.denommetadataKeeper.CreateDenomMetadata(ctx, m); err != nil {
		return fmt.Errorf("create denom metadata: %w", err)
	}

	k.Logger(ctx).Info("registered denom metadata for IBC token", "rollappID", rollappID, "denom", ibcDenom)
	return nil
}
