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
		k.Logger(ctx).Error("genesis event already happened")
		// panic(errors.New("genesis event already happened - it shouldn't have")) TODO:
	}

	rollapp.GenesisState.GenesisEventHappened = true
	k.SetRollapp(ctx, rollapp)

	return nil
}

func (k Keeper) RegisterOneDenomMetadata(ctx sdktypes.Context, m banktypes.Metadata, rollappID, channelID string) error {
	trace := utils.GetForeignDenomTrace(channelID, m.Base)

	k.transferKeeper.SetDenomTrace(ctx, trace)

	ibcDenom := trace.IBCDenom()

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

	if err := m.Validate(); err != nil {
		// TODO: errorsmod with nice wrapping
		return fmt.Errorf("invalid denom metadata on genesis event: %w", err)
	}

	k.bankKeeper.SetDenomMetaData(ctx, m)

	k.Logger(ctx).Info("Registered denom metadata for IBC token.", "rollappID", rollappID, "denom", ibcDenom)
	return nil
}
