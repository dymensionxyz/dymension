package transfersenabled

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transferTypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	"github.com/dymensionxyz/dymension/v3/utils/gerr"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

type GetRollapp func(ctx sdk.Context, rollappId string) (val types.Rollapp, found bool)

type Decorator struct {
	getRollapp GetRollapp
}

func NewDecorator(getRollapp GetRollapp) *Decorator {
	return &Decorator{
		getRollapp: getRollapp,
	}
}

func (h Decorator) transfersEnabled(ctx sdk.Context, transfer *transferTypes.MsgTransfer) (bool, error) {
	/*
		TODO:
		need to get the intended rollapp and check if the transfers are enabled
	*/
	return false, nil
}

func (h Decorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	for _, msg := range tx.GetMsgs() {
		if m, ok := msg.(*transferTypes.MsgTransfer); ok {
			ok, err := h.transfersEnabled(ctx, m)
			if err != nil {
				return ctx, errorsmod.Wrap(err, "transfer genesis: transfers enabled")
			}
			if !ok {
				return ctx, errorsmod.Wrap(gerr.ErrFailedPrecondition, "transfers for rollapp are disabled")
			}
		}
	}

	return next(ctx, tx, simulate)
}
