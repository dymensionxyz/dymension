package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctypes "github.com/cosmos/ibc-go/v6/modules/light-clients/07-tendermint/types"
)

func (k Keeper) ExtractRollappIDFromChannel(ctx sdk.Context, portID string, channelID string) (string, error) {
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
