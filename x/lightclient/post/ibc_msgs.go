package post

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/keeper"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
)

var _ sdk.PostDecorator = IBCMessagesDecorator{}

type IBCMessagesDecorator struct {
	ibcClientKeeper   types.IBCClientKeeperExpected
	lightClientKeeper keeper.Keeper
}

func NewIBCMessagesDecorator(k keeper.Keeper, ibcK types.IBCClientKeeperExpected) IBCMessagesDecorator {
	return IBCMessagesDecorator{
		ibcClientKeeper:   ibcK,
		lightClientKeeper: k,
	}
}

func (i IBCMessagesDecorator) PostHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, success bool, next sdk.PostHandler) (newCtx sdk.Context, err error) {
	msgs := tx.GetMsgs()
	for _, m := range msgs {
		switch msg := m.(type) {
		case *ibcclienttypes.MsgCreateClient:
			i.HandleMsgCreateClient(ctx, msg, success)
		default:
			continue
		}
	}
	return next(ctx, tx, simulate, success)
}
