package ante

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	"github.com/dymensionxyz/dymension/v3/x/lightclient/keeper"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

var _ sdk.AnteDecorator = IBCMessagesDecorator{}

type RollappKeeperExpected interface {
	GetRollapp(ctx sdk.Context, rollappId string) (val rollapptypes.Rollapp, found bool)
	FindStateInfoByHeight(ctx sdk.Context, rollappId string, height uint64) (rollapptypes.StateInfo, error)
}

type IBCMessagesDecorator struct {
	ibcKeeper         ibckeeper.Keeper
	rollappKeeper     RollappKeeperExpected
	lightClientKeeper keeper.Keeper
}

func NewIBCMessagesDecorator() IBCMessagesDecorator {
	return IBCMessagesDecorator{}
}

// AnteHandle implements types.AnteDecorator.
func (i IBCMessagesDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	msgs := tx.GetMsgs()
	for _, m := range msgs {
		switch msg := m.(type) {
		case *ibcclienttypes.MsgCreateClient:
			i.HandleMsgCreateClient(ctx, msg)
		case *ibcclienttypes.MsgUpdateClient:
			{
			}
		case *ibcclienttypes.MsgSubmitMisbehaviour:
			{
			}
		default:
			continue
		}
	}
	return next(ctx, tx, simulate)
}
