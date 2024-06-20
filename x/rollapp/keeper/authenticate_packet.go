package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	uibc "github.com/dymensionxyz/dymension/v3/utils/ibc"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

/*
TODO: : instead of calling these all the time in every middleware, we could do it once and use
	context https://github.com/dymensionxyz/dymension/issues/914
*/

// GetValidTransfer takes a packet, ensures it is a (basic) validated fungible token packet, and gets the chain id.
// If the channel chain id is also a rollapp id, we check that the canonical channel id we have saved for that rollapp
// agrees is indeed the channel we are receiving from. In this way, we stop anyone from pretending to be the RA. (Assuming
// that the mechanism for setting the canonical channel in the first place is correct).
func (k Keeper) GetValidTransfer(
	ctx sdk.Context,
	packetData []byte,
	raPortOnHub, raChanOnHub string,
) (data types.TransferData, err error) {
	if err = transfertypes.ModuleCdc.UnmarshalJSON(packetData, &data.FungibleTokenPacketData); err != nil {
		err = errorsmod.Wrap(err, "unmarshal transfer data")
		return
	}

	if err = data.ValidateBasic(); err != nil {
		err = errorsmod.Wrap(err, "validate basic")
		return
	}

	ra, err := k.getRollappByPortChan(ctx, raPortOnHub, raChanOnHub)
	if errorsmod.IsOf(err, errRollappNotFound) {
		// no problem, it corresponds to a regular non-rollapp chain
		err = nil
		return
	}
	if err != nil {
		err = errorsmod.Wrap(err, "get rollapp id")
		return
	}

	data.Rollapp = ra

	return
}

var errRollappNotFound = errorsmod.Wrap(gerrc.ErrNotFound, "rollapp")

// getRollappByPortChan returns the rollapp id that a packet came from, if we are certain
// that the packet came from that rollapp. That means that the canonical channel
// has already been set.
func (k Keeper) getRollappByPortChan(ctx sdk.Context,
	raPortOnHub, raChanOnHub string,
) (*types.Rollapp, error) {
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
	chainID, err := uibc.ChainIDFromPortChannel(ctx, k.channelKeeper.GetChannelClientState, raPortOnHub, raChanOnHub)
	if err != nil {
		return nil, errorsmod.Wrap(err, "chain id from port and channel")
	}
	rollapp, ok := k.GetRollapp(ctx, chainID)
	if !ok {
		return nil, errorsmod.Wrapf(errRollappNotFound, "chain id: %s: port: %s: channel: %s", chainID, raPortOnHub, raChanOnHub)
	}
	if rollapp.ChannelId == "" {
		return nil, errorsmod.Wrapf(gerrc.ErrFailedPrecondition, "rollapp canonical channel mapping has not been set: %s", chainID)
	}

	if rollapp.ChannelId != raChanOnHub {
		return nil, errorsmod.Wrapf(
			gerrc.ErrInvalidArgument,
			"packet destination channel id mismatch: expect: %s: got: %s", rollapp.ChannelId, raChanOnHub,
		)
	}
	return &rollapp, nil
}
