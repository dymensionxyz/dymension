package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibctypes "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// ExtractRollappIDAndTransferPacketFromData extracts the rollapp ID and fungible token from the packet data
// Returns an empty string if the rollapp is not found
func (k Keeper) ExtractRollappIDAndTransferPacketFromData(
	ctx sdk.Context,
	data []byte,
	rollappPortOnHub string,
	rollappChannelOnHub string,
) (string, *transfertypes.FungibleTokenPacketData, error) {
	// no-op if the packet is not a fungible token packet
	packet := new(transfertypes.FungibleTokenPacketData)
	if err := types.ModuleCdc.UnmarshalJSON(data, packet); err != nil {
		return "", packet, errorsmod.Wrapf(err, "unmarshal packet data")
	}

	rollapp, err := k.ExtractRollappFromChannel(ctx, rollappPortOnHub, rollappChannelOnHub)
	if err != nil {
		return "", packet, err
	}

	if rollapp == nil {
		return "", packet, nil
	}

	return rollapp.RollappId, packet, nil
}

// ExtractRollappFromChannel extracts the rollapp from the IBC port and channel.
// Returns nil if the rollapp is not found.
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

// ExtractChainIDFromChannel extracts the chain ID from the channel
func (k Keeper) ExtractChainIDFromChannel(ctx sdk.Context, portID string, channelID string) (string, error) {
	_, clientState, err := k.channelKeeper.GetChannelClientState(ctx, portID, channelID)
	if err != nil {
		return "", fmt.Errorf("extract clientID from channel: %w", err)
	}

	tmClientState, ok := clientState.(*ibctypes.ClientState)
	if !ok {
		return "", nil
	}

	return tmClientState.ChainId, nil
}
