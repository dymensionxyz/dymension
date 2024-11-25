package genesisbridge

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transferTypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// TODO: refactor this to use ICS4 wrapper similar to the RDK
// (https://github.com/dymensionxyz/dymension/issues/957)

type ChannelKeeper interface {
	GetChannelClientState(ctx sdk.Context, portID, channelID string) (string, exported.ClientState, error) // implemented by ibc channel keeper
}

// TransferEnabledDecorator only allows ibc transfers to a rollapp if that rollapp has finished
// the genesis bridge protocol.
type TransferEnabledDecorator struct {
	rollappK              RollappKeeperMinimal
	getChannelClientState ChannelKeeper
}

func NewTransferEnabledDecorator(rollappK RollappKeeperMinimal, getChannelClientState ChannelKeeper) *TransferEnabledDecorator {
	return &TransferEnabledDecorator{
		rollappK:              rollappK,
		getChannelClientState: getChannelClientState,
	}
}

func (h TransferEnabledDecorator) transfersEnabled(ctx sdk.Context, transfer *transferTypes.MsgTransfer) (bool, error) {
	ra, err := h.rollappK.GetRollappByPortChan(ctx, transfer.SourcePort, transfer.SourceChannel)
	if err != nil {
		if errorsmod.IsOf(err, types.ErrRollappNotFound) {
			// Two cases
			// 1. Non rollapp
			//    Transfers are enabled
			// 2. It is for rollapp but the light client of this transfer is not yet canonical or will never
			//    be marked canonical: for a correct rollapp the transfer channel is only created after it's
			//    marked canonical, so this transfer corresponds to a not-relevant channel.
			//    Note: IBC prevents sending to a channel which is not OPEN, which prevents making the transfer
			//    for a channel before it is marked canonical in the onOpenAck hook.
			return true, nil
		}
		return false, errorsmod.Wrap(err, "rollapp by port chan")
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
