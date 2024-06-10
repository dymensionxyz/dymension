package ibc

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
	ibctmtypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"
	"github.com/dymensionxyz/dymension/v3/utils/gerr"
)

type GetChannelClientState func(ctx sdk.Context, portID, channelID string) (string, exported.ClientState, error)

func ChainIDFromPortChannel(
	ctx sdk.Context,
	getChannelClientState GetChannelClientState,
	portID string,
	channelID string,
) (string, error) {
	_, state, err := getChannelClientState(ctx, portID, channelID)
	if err != nil {
		return "", errorsmod.Wrap(err, "get channel client state")
	}

	tmState, ok := state.(*ibctmtypes.ClientState)
	if !ok {
		return "", errorsmod.Wrap(gerr.ErrInvalidArgument, "cast client state to tmtypes client state")
	}

	return tmState.ChainId, nil
}
