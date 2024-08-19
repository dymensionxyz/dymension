package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/keeper"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
)

var _ sdk.AnteDecorator = IBCMessagesDecorator{}

type IBCMessagesDecorator struct {
	ibcKeeper         ibckeeper.Keeper
	rollappKeeper     types.RollappKeeperExpected
	lightClientKeeper keeper.Keeper
}

func NewIBCMessagesDecorator(k keeper.Keeper, ibcK ibckeeper.Keeper, rk types.RollappKeeperExpected) IBCMessagesDecorator {
	return IBCMessagesDecorator{
		ibcKeeper:         ibcK,
		rollappKeeper:     rk,
		lightClientKeeper: k,
	}
}

func (i IBCMessagesDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	msgs := tx.GetMsgs()
	for _, m := range msgs {
		switch msg := m.(type) {
		case *ibcclienttypes.MsgCreateClient:
			i.HandleMsgCreateClient(ctx, msg)
		case *ibcclienttypes.MsgSubmitMisbehaviour:
			if err := i.HandleMsgSubmitMisbehaviour(ctx, msg); err != nil {
				return ctx, err
			}
		case *ibcclienttypes.MsgUpdateClient:
			if err := i.HandleMsgUpdateClient(ctx, msg); err != nil {
				return ctx, err
			}
		case *ibcchanneltypes.MsgChannelOpenAck:
			if err := i.HandleMsgChannelOpenAck(ctx, msg); err != nil {
				return ctx, err
			}
		default:
			continue
		}
	}
	return next(ctx, tx, simulate)
}
