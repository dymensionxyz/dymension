package keeper

import (
	"github.com/dymensionxyz/dymension/v3/utils/gerr"
	uibc "github.com/dymensionxyz/dymension/v3/utils/ibc"
	"github.com/dymensionxyz/dymension/v3/x/delayedack/types"

	errorsmod "cosmossdk.io/errors"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetValidTransfer takes a packet, ensures it is a (basic) validated fungible token packet, and gets the chain id.
// If the channel chain id is also a rollapp id, we check that the canonical channel id we have saved for that rollapp
// agrees is indeed the channel we are receiving from. In this way, we stop anyone from pretending to be the RA. (Assuming
// that the mechanism for setting the canonical channel in the first place is correct).
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

	rollappID, err := k.getRollappID(ctx, packet)
	if errorsmod.IsOf(err, ErrRollappNotFound) {
		// no problem, it corresponds to a regular non-rollapp chain
		return
	}
	if err != nil {
		err = errorsmod.Wrap(err, "get rollapp id")
		return
	}

	data.RollappID = rollappID

	return
}

var ErrRollappNotFound = errorsmod.Wrap(gerr.ErrNotFound, "rollapp")

// getRollappID returns the rollapp id that a packet came from, if we are certain
// that the packet came from that rollapp. That means that the canonical channel
// has already been set.
func (k Keeper) getRollappID(ctx sdk.Context,
	packet channeltypes.Packet,
) (string, error) {
	/*
		TODO:
			There is an open issue of how we go about making sure that the packet really came from the rollapp, and once we know that it came
			from the rollapp, also how we deal with fraud from the sequencer
			See https://github.com/dymensionxyz/research/issues/242 for info
			See
				https://github.com/dymensionxyz/dymension/blob/8734e239483bb6290de6d01c196da35fa033e160/x/delayedack/keeper/authenticate_packet.go#L100-L204
				https://github.com/dymensionxyz/dymension/blob/8734e239483bb6290de6d01c196da35fa033e160/x/delayedack/keeper/authenticate_packet.go#L100-L204
				https://github.com/dymensionxyz/dymension/blob/a74ffb0cec00768bbb8dbe3fd6413e66388010d3/x/delayedack/keeper/keeper.go#L98-L107
				https://github.com/dymensionxyz/dymension/blob/986d51ccd4807d514c91b3a147ac1b8ce5b590a1/x/delayedack/keeper/authenticate_packet.go#L47-L59
				for the old implementations of checks
	*/
	chainID, err := uibc.ChainIDFromPortChannel(ctx, k.channelKeeper.GetChannelClientState, packet.GetDestPort(), packet.GetDestChannel())
	if err != nil {
		return "", errorsmod.Wrap(err, "chain id from port and channel")
	}
	rollapp, ok := k.rollappKeeper.GetRollapp(ctx, chainID)
	if !ok {
		return "", ErrRollappNotFound
	}
	if rollapp.ChannelId == "" {
		return "", errorsmod.Wrapf(gerr.ErrFailedPrecondition, "rollapp canonical channel mapping has not been set: %s", chainID)
	}

	if rollapp.ChannelId != packet.DestinationChannel {
		return "", errorsmod.Wrapf(
			gerr.ErrInvalidArgument,
			"packet destination channel id mismatch: expect: %s: got: %s", rollapp.ChannelId, packet.DestinationChannel,
		)
	}
	return rollapp.ChannelId, nil
}
