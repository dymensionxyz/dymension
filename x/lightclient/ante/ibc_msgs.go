package ante

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibcchanneltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/keeper"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
)

var _ sdk.AnteDecorator = IBCMessagesDecorator{}

type IBCMessagesDecorator struct {
	ibcClientKeeper  types.IBCClientKeeperExpected
	ibcChannelKeeper types.IBCChannelKeeperExpected
	raK              types.RollappKeeperExpected
	k                keeper.Keeper
}

func NewIBCMessagesDecorator(
	k keeper.Keeper,
	ibcClient types.IBCClientKeeperExpected,
	ibcChannel types.IBCChannelKeeperExpected,
	rk types.RollappKeeperExpected,
) IBCMessagesDecorator {
	return IBCMessagesDecorator{
		ibcClientKeeper:  ibcClient,
		ibcChannelKeeper: ibcChannel,
		raK:              rk,
		k:                k,
	}
}

func (i IBCMessagesDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	msgs := tx.GetMsgs()
	for _, m := range msgs {
		// TODO: need to handle authz etc
		switch msg := m.(type) {
		case *ibcclienttypes.MsgSubmitMisbehaviour:
			if err := i.HandleMsgSubmitMisbehaviour(ctx, msg); err != nil {
				return ctx, errorsmod.Wrap(err, "handle MsgSubmitMisbehaviour")
			}
		case *ibcclienttypes.MsgUpdateClient:
			if err := i.HandleMsgUpdateClient(ctx, msg); err != nil {
				return ctx, errorsmod.Wrap(err, "handle MsgUpdateClient")
			}
		case *ibcchanneltypes.MsgChannelOpenAck:
			if err := i.HandleMsgChannelOpenAck(ctx, msg); err != nil {
				return ctx, errorsmod.Wrap(err, "handle MsgChannelOpenAck")
			}
		default:
			continue
		}
	}
	return next(ctx, tx, simulate)
}
