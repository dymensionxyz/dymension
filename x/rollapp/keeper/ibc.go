package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcdymtypes "github.com/cosmos/ibc-go/v6/modules/light-clients/01-dymint/types"
)

func (k Keeper) ExtractRollappIDFromChannel(ctx sdk.Context, portID string, channelID string) (string, error) {
	_, clientState, err := k.channelKeeper.GetChannelClientState(ctx, portID, channelID)
	if err != nil {
		return "", fmt.Errorf("failed to extract clientID from channel: %w", err)
	}

	tmClientState, ok := clientState.(*ibcdymtypes.ClientState)
	if !ok {
		return "", nil
	}

	return tmClientState.ChainId, nil
}
