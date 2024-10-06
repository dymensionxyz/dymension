package genesisbridge

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uibc"

	transferTypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

// FIXME: refactor this to use the ibc module's channel keeper
// (https://github.com/dymensionxyz/dymension/issues/957)

type GetRollapp func(ctx sdk.Context, rollappId string) (val types.Rollapp, found bool)

type ChannelKeeper interface {
	GetChannelClientState(ctx sdk.Context, portID, channelID string) (string, exported.ClientState, error) // implemented by ibc channel keeper
}

// TransferEnabledDecorator only allows ibc transfers to a rollapp if that rollapp has finished
// the genesis bridge protocol.
type TransferEnabledDecorator struct {
	getRollapp            GetRollapp
	getChannelClientState ChannelKeeper
}

func NewTransferEnabledDecorator(getRollapp GetRollapp, getChannelClientState ChannelKeeper) *TransferEnabledDecorator {
	return &TransferEnabledDecorator{
		getRollapp:            getRollapp,
		getChannelClientState: getChannelClientState,
	}
}

func (h TransferEnabledDecorator) transfersEnabled(ctx sdk.Context, transfer *transferTypes.MsgTransfer) (bool, error) {
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

// AnteHandle will return an error if the tx contains an ibc transfer message to a rollapp that has not finished the transfer genesis protocol.
func (h TransferEnabledDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (sdk.Context, error) {
	for _, msg := range tx.GetMsgs() {
		typeURL := sdk.MsgTypeURL(msg)
		if typeURL == sdk.MsgTypeURL(&transferTypes.MsgTransfer{}) {
			m, ok := msg.(*transferTypes.MsgTransfer)
			if !ok {
				return ctx, errorsmod.Wrap(gerrc.ErrUnknown, "type url matched transfer type url but could not type cast")
			}
			ok, err := h.transfersEnabled(ctx, m)
			if err != nil {
				return ctx, errorsmod.Wrap(err, "transfer genesis: transfers enabled")
			}
			if !ok {
				return ctx, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "transfers to/from rollapp are disabled")
			}
		}
	}

	return next(ctx, tx, simulate)
}
