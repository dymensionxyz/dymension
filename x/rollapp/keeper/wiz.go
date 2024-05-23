package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"

	sdktypes "github.com/cosmos/cosmos-sdk/types"
	tenderminttypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

func (k Keeper) TriggerGen(goCtx context.Context, msg types.GenParams) error {
	/*
		What does this do?

		Get the client for the channel, and check the tendermint chain id matches the rollapp id
		Sets the rollapp channel id
	*/

	ctx := sdktypes.UnwrapSDKContext(goCtx)

	// NOTE: whitelist check removed here

	// Get the rollapp
	rollapp, found := k.GetRollapp(ctx, msg.RollappID)
	if !found {
		return types.ErrUnknownRollappID
	}

	// Validate it hasn't been triggered yet
	if rollapp.GenesisState.GenesisEventHappened {
		return types.ErrGenesisEventAlreadyTriggered
	}

	// Get the channel and validate it's connected client chain is the same as the rollapp's
	_, clientState, err := k.channelKeeper.GetChannelClientState(ctx, "transfer", msg.ChannelID)
	if err != nil {
		return fmt.Errorf("get channel client state: %w", err)
	}
	tmClientState, ok := clientState.(*tenderminttypes.ClientState)
	if !ok {
		return errorsmod.Wrapf(types.ErrInvalidGenesisChannelId, "expected tendermint client state, got %T", clientState)
	}
	if tmClientState.GetChainID() != msg.RollappID {
		return errorsmod.Wrapf(types.ErrInvalidGenesisChannelId, "channel connected to wrong chain: channel: %s: got: %s: expect: %s",
			msg.ChannelID, tmClientState.GetChainID(), msg.RollappID)
	}

	// Update the rollapp with the channelID and trigger the genesis event
	rollapp.ChannelId = msg.ChannelID

	if err := k.registerDenomMetadata(ctx, rollapp); err != nil {
		return errorsmod.Wrapf(types.ErrRegisterDenomMetadataFailed, "register denom metadata: %s", err)
	}

	if err := k.mintRollappGenesisTokens(ctx, rollapp); err != nil {
		return errorsmod.Wrapf(types.ErrMintTokensFailed, "mint rollapp genesis tokens: %s", err)
	}

	rollapp.GenesisState.GenesisEventHappened = true
	k.SetRollapp(ctx, rollapp)

	return nil
}
