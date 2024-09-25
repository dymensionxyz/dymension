package keeper

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
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

// getRollappByPortChan will get the rollapp for a transfer
// if the transfer did not original from a rollapp, will return rollapp not found error
// if the transfer did originate from a rollapp, but on the wrong channel, returns error
//
// in order to allow rollapp and non rollapps to have the same chain ID, the (possible)
// rollapp is looked up by light client ID rather than chain ID. That requires the canonical
// light client for the rollapp to have been set. That should always be the case for
// correctly operated rollapps.
func (k Keeper) getRollappByPortChan(ctx sdk.Context,
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
		return nil, errorsmod.Wrapf(errRollappNotFound, "client id: %s: port: %s: channel: %s", clientID, raPortOnHub, raChanOnHub)
	}
	rollapp, ok := k.GetRollapp(ctx, chainID)
	if !ok {
		return nil, errorsmod.Wrap(gerrc.ErrInternal, "have canonical client id but rollapp not found")
	}
	if rollapp.ChannelId == "" {
		return nil, errorsmod.Wrap(gerrc.ErrInternal, "canonical client for rollapp is set, but canonical channel is missing")
	}
	if rollapp.ChannelId != raChanOnHub {
		return nil, errorsmod.Wrapf(
			gerrc.ErrInvalidArgument,
			"transfer from rollapp is not on canonical channel, packet destination channel id mismatch: expect: %s: got: %s", rollapp.ChannelId, raChanOnHub,
		)
	}
	return &rollapp, nil
}
