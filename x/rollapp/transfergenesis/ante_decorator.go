package transfergenesis

import (
	uibc "github.com/dymensionxyz/dymension/v3/utils/ibc"

	errorsmod "cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	transferTypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	"github.com/dymensionxyz/dymension/v3/utils/gerr"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

type GetRollapp func(ctx sdk.Context, rollappId string) (val types.Rollapp, found bool)

type Decorator struct {
	getRollapp            GetRollapp
	getChannelClientState uibc.GetChannelClientState
}

func NewDecorator(getRollapp GetRollapp, getChannelClientState uibc.GetChannelClientState) *Decorator {
	return &Decorator{
		getRollapp:            getRollapp,
		getChannelClientState: getChannelClientState,
	}
}

func (h Decorator) transfersEnabled(ctx sdk.Context, transfer *transferTypes.MsgTransfer) (bool, error) {
	chainID, err := uibc.ChainIDFromPortChannel(ctx, h.getChannelClientState, transfer.SourcePort, transfer.SourceChannel)
	if err != nil {
		return false, errorsmod.Wrap(err, "chain id from port channel")
	}
	ra, ok := h.getRollapp(ctx, chainID)
	if !ok {
		return true, nil
	}
	return ra.GenesisState.TransfersEnabled, nil
}

func (h Decorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	for _, msg := range tx.GetMsgs() {
		typeURL := sdk.MsgTypeURL(msg)
		if typeURL == sdk.MsgTypeURL(&transferTypes.MsgTransfer{}) {
			m, ok := msg.(*transferTypes.MsgTransfer)
			if !ok {
				return ctx, errorsmod.Wrap(gerr.ErrUnknown, "type url matched transfer type url but could not type cast")
			}
			ok, err := h.transfersEnabled(ctx, m)
			if err != nil {
				return ctx, errorsmod.Wrap(err, "transfer genesis: transfers enabled")
			}
			if !ok {
				return ctx, errorsmod.Wrap(gerr.ErrFailedPrecondition, "transfers to/from rollapp are disabled")
			}
		}
	}

	return next(ctx, tx, simulate)
}
