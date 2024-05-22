package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	porttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
)

type IBCMiddleware struct {
	porttypes.IBCModule
	keeper *Keeper
}

func NewIBCMiddleware(next porttypes.IBCModule, keeper *Keeper) *IBCMiddleware {
	return &IBCMiddleware{
		IBCModule: next,
		keeper:    keeper,
	}
}

func (im IBCMiddleware) OnChanOpenConfirm(
	ctx sdk.Context,
	portID,
	channelID string,
) error {
	/*
		TODO:
	*/
	return im.IBCModule.OnChanOpenConfirm(ctx, portID, channelID)
}
