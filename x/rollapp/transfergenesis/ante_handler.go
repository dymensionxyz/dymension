package transfergenesis

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	"github.com/dymensionxyz/dymension/v3/utils/gerr"
)

type Decorator struct {
	// Keep any required keepers or state here
}

func (h Decorator) transfersEnabled(ctx sdk.Context, transfer *ibctransfertypes.MsgTransfer) (bool, error) {
}

func (h Decorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	for _, msg := range tx.GetMsgs() {
		if m, ok := msg.(*ibctransfertypes.MsgTransfer); ok {
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
