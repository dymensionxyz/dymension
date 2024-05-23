package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	ibctypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// ExtractRollappAndTransferPacketFromData extracts the rollapp and fungible token from the packet data
func (k Keeper) ExtractRollappAndTransferPacketFromData(
	ctx sdk.Context,
	data []byte,
	rollappPortOnHub string,
	rollappChannelOnHub string,
) (*types.Rollapp, *transfertypes.FungibleTokenPacketData, error) {
	// no-op if the packet is not a fungible token packet
	packet := new(transfertypes.FungibleTokenPacketData)
	if err := types.ModuleCdc.UnmarshalJSON(data, packet); err != nil {
		return nil, nil, errorsmod.Wrapf(err, "failed to unmarshal packet data")
	}

	rollapp, err := k.ExtractRollappFromChannel(ctx, rollappPortOnHub, rollappChannelOnHub)
	if err != nil {
		return nil, nil, err
	}

	return rollapp, packet, nil
}

// ExtractRollappFromChannel extracts the rollapp from the IBC port and channel
func (k Keeper) ExtractRollappFromChannel(
	ctx sdk.Context,
	rollappPortOnHub string,
	rollappChannelOnHub string,
) (*types.Rollapp, error) {
	// Check if the packet is destined for a rollapp
	chainID, err := k.ExtractChainIDFromChannel(ctx, rollappPortOnHub, rollappChannelOnHub)
	if err != nil {
		return nil, err
	}

	rollapp, found := k.GetRollapp(ctx, chainID)
	if !found {
		return nil, nil
	}

	if rollapp.ChannelId == "" {
		return nil, errorsmod.Wrapf(types.ErrGenesisEventNotTriggered, "empty channel id: rollap id: %s", chainID)
	}
	// check if the channelID matches the rollappID's channelID
	if rollapp.ChannelId != rollappChannelOnHub {
		return nil, errorsmod.Wrapf(
			types.ErrMismatchedChannelID,
			"channel id mismatch: expect: %s: got: %s", rollapp.ChannelId, rollappChannelOnHub,
		)
	}

	return &rollapp, nil
}

func (k Keeper) ExtractChainIDFromChannel(ctx sdk.Context, portID string, channelID string) (string, error) {
	_, clientState, err := k.channelKeeper.GetChannelClientState(ctx, portID, channelID)
	if err != nil {
		return "", fmt.Errorf("failed to extract clientID from channel: %w", err)
	}

	tmClientState, ok := clientState.(*ibctypes.ClientState)
	if !ok {
		return "", nil
	}

	return tmClientState.ChainId, nil
}
