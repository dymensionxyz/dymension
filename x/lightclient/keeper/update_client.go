package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/ante"
)

// wraps the normal ibc client keeper update client message but routes it through our ante
// Now we have two ways to update: direct through normal pathway or here, which is messy.
// We can improve in SDK v0.52+ with pre/post message hooks.
func (m msgServer) UpdateClient(goCtx context.Context, msg *clienttypes.MsgUpdateClient) (*clienttypes.MsgUpdateClientResponse, error) {

	ctx := sdk.UnwrapSDKContext(goCtx)

	d := ante.NewIBCMessagesDecorator(*m.Keeper, m.ibcClientKeeper, m.ibcChannelK, m.rollappKeeper)

	err := d.HandleMsgUpdateClient(ctx, msg)

	if err != nil {
		return nil, err
	}

	return m.ibcKeeper.UpdateClient(ctx, msg)
}
