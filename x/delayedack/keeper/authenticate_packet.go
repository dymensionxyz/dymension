package keeper

import (
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"

	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"

	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	ibctmtypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetValidTransfer takes a packet, ensures it is a (basic) validated fungible token packet, and gets the chain id.
// If the channel chain id is also a rollapp id, we check that the canonical channel id we have saved for that rollapp
// agrees is indeed the channel we are receiving from.
// If packet HAS come from the canonical channel, we also
func (k Keeper) GetValidTransfer(
	ctx sdk.Context,
	packet channeltypes.Packet,
) (data types.TransferData, err error) {
	if err = transfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		err = errorsmod.Wrap(err, "unmarshal transfer data")
		return
	}

	if err = data.ValidateBasic(); err != nil {
		err = errorsmod.Wrap(err, "validate basic")
		return
	}

	chainID, err := k.chainIDFromPortChannel(ctx, packet.DestinationPort, packet.DestinationChannel)
	if err != nil {
		err = errorsmod.Wrap(err, "chain id from port and channel")
		return
	}

	rollapp, ok := k.rollappKeeper.GetRollapp(ctx, chainID)
	if !ok {
		// no problem, it corresponds to a regular non-rollapp chain
		return
	}

	data.RollappID = chainID
	if rollapp.ChannelId == "" {
		err = errorsmod.Wrapf(rollapptypes.ErrGenesisEventNotTriggered, "empty channel id: rollap id: %s", chainID)
		return
	}

	if rollapp.ChannelId != packet.DestinationChannel {
		err = errorsmod.Wrapf(
			rollapptypes.ErrMismatchedChannelID,
			"channel id mismatch: expect: %s: got: %s", rollapp.ChannelId, packet.DestinationChannel,
		)
		return
	}

	/*
		TODO:
			There is an open issue of how we go about making sure that the packet really came from the rollapp, and once we know that it came
			from the rollapp, also how we deal with fraud from the sequencer
			See https://github.com/dymensionxyz/research/issues/242 for info
			See
				https://github.com/dymensionxyz/dymension/blob/8734e239483bb6290de6d01c196da35fa033e160/x/delayedack/keeper/authenticate_packet.go#L100-L204
				https://github.com/dymensionxyz/dymension/blob/8734e239483bb6290de6d01c196da35fa033e160/x/delayedack/keeper/authenticate_packet.go#L100-L204
				for the old implementation
	*/

	return
}

func (k Keeper) chainIDFromPortChannel(ctx sdk.Context, portID string, channelID string) (string, error) {
	_, state, err := k.channelKeeper.GetChannelClientState(ctx, portID, channelID)
	if err != nil {
		return "", errorsmod.Wrap(err, "get channel client state")
	}

	tmState, ok := state.(*ibctmtypes.ClientState)
	if !ok {
		return "", nil
	}

	return tmState.ChainId, nil
}
