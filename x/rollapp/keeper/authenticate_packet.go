package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
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

	ra, err := k.GetRollappByCanonicalChan(ctx, raPortOnHub, raChanOnHub)
	if errorsmod.IsOf(err, types.ErrRollappNotFound) {
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

// GetRollappByCanonicalChan retrieves the rollapp associated with a specific canonical channel.
// It checks if the rollapp's canonical channel ID matches the provided channel ID.
// This function ensures that the rollapp has a canonical channel set and that the packet
// is being transferred on the correct channel, which is essential after the genesis bridge
// has been opened. It returns an error if the rollapp does not have a canonical channel set
// or if the packet is not on the canonical channel.
func (k Keeper) GetRollappByCanonicalChan(ctx sdk.Context,
	raPortOnHub, raChanOnHub string,
) (*types.Rollapp, error) {
	rollapp, err := k.GetRollappByPortChan(ctx, raPortOnHub, raChanOnHub)
	if err != nil {
		return nil, err
	}

	// if canonical channel is not set, return error
	if rollapp.ChannelId == "" {
		return nil, errorsmod.Wrap(gerrc.ErrInternal, "canonical client for rollapp is set, but canonical channel is missing")
	}

	// if the channel id does not match, return error
	if rollapp.ChannelId != raChanOnHub {
		return nil, errorsmod.Wrapf(
			gerrc.ErrInvalidArgument,
			"transfer from rollapp is not on canonical channel, packet destination channel id mismatch: expect: %s: got: %s", rollapp.ChannelId, raChanOnHub,
		)
	}

	return rollapp, nil
}

// GetRollappByPortChan retrieves the rollapp for a transfer based on the port and channel.
// This function checks for any channel of a rollapp, not necessarily the canonical one.
// It uses the light client ID to find the rollapp, which means the canonical light client
// must be set for the rollapp. This is suitable for scenarios where the genesis bridge
// has not yet been opened. It returns an error if the rollapp is not found or if the
// rollapp does not have a canonical client set.
func (k Keeper) GetRollappByPortChan(ctx sdk.Context,
	raPortOnHub, raChanOnHub string,
) (*types.Rollapp, error) {
	clientID, _, err := k.channelKeeper.GetChannelClientState(ctx, raPortOnHub, raChanOnHub)
	if err != nil {
		return nil, errorsmod.Wrap(err, "get chan client state")
	}
	chainID, ok := k.canonicalClientKeeper.GetRollappForClientID(ctx, clientID)
	if !ok {
		// non rollapp case. Note, we cannot differentiate the case where the transfer is not from a rollapp, or it is from a rollapp
		// but that rollapp has (incorrectly) not got a canonical client
		return nil, errorsmod.Wrapf(types.ErrRollappNotFound, "client id: %s: port: %s: channel: %s", clientID, raPortOnHub, raChanOnHub)
	}
	rollapp, ok := k.GetRollapp(ctx, chainID)
	if !ok {
		return nil, errorsmod.Wrap(gerrc.ErrInternal, "have canonical client id but rollapp not found")
	}

	return &rollapp, nil
}
