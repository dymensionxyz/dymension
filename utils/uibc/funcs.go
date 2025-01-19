package uibc

import (
	"strings"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibctmtypes "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func IsIBCDenom(denom string) bool {
	return strings.HasPrefix(denom, "ibc/")
}

const transferPort = "transfer"

func GetForeignDenomTrace(channelId string, denom string) transfertypes.DenomTrace {
	sourcePrefix := transfertypes.GetDenomPrefix(transferPort, channelId)
	// NOTE: sourcePrefix contains the trailing "/"
	prefixedDenom := sourcePrefix + denom
	// construct the denomination trace from the full raw denomination
	return transfertypes.ParseDenomTrace(prefixedDenom)
}

// ChainIDFromPortChannelKeeper is a dependency for ChainIDFromPortChannel(
// Deprecated: ChainIDFromPortChannel is deprecated.
type ChainIDFromPortChannelKeeper interface {
	GetChannelClientState(ctx sdk.Context, portID, channelID string) (string, exported.ClientState, error) // implemented by ibc channel keeper
}

// ChainIDFromPortChannel returns the chain ID associated to the light client for the channel
// Deprecated: this method can still be used if chain ID is needed, but the chain ID of the
// light client is not a reliable way to identify a rollapp, since a rollapp can have the same
// chain ID as a non rollapp chain connected to the hub.
func ChainIDFromPortChannel(
	ctx sdk.Context,
	keeper ChainIDFromPortChannelKeeper,
	portID string,
	channelID string,
) (string, error) {
	_, state, err := keeper.GetChannelClientState(ctx, portID, channelID)
	if err != nil {
		return "", errorsmod.Wrap(err, "get channel client state")
	}

	tmState, ok := state.(*ibctmtypes.ClientState)
	if !ok {
		return "", errorsmod.WithType(gerrc.ErrInvalidArgument, state)
	}

	return tmState.ChainId, nil
}
