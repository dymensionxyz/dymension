package genesisbridge

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	gerrc "github.com/dymensionxyz/gerr-cosmos/gerrc"
)

const (
	// HubRecipient is the address of `x/rollapp` module's account on the hub chain.
	HubRecipient = "dym1mk7pw34ypusacm29m92zshgxee3yreums8avur"
)

// HandleGenesisTransfer handles the genesis transfer packet, if present, and expected.
// We assume that genesis info is already validated, and the genesis transfer is expected to fulfill the genesis info.
func (w IBCModule) handleGenesisTransfer(ctx sdk.Context, ra types.Rollapp, packet channeltypes.Packet, gTransfer *transfertypes.FungibleTokenPacketData) error {
	// check if required or expected
	required := ra.GenesisInfo.GenesisAccounts != nil
	// required but not present
	if required && gTransfer == nil {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "genesis transfer required")
	}
	// not required but present
	if !required && gTransfer != nil {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "genesis transfer not expected")
	}
	if gTransfer == nil {
		return nil
	}

	// validate the receiver
	if gTransfer.Receiver != HubRecipient {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "receiver mismatch")
	}

	// validate that the transfer amount matches the expected amount, which is the sum of all genesis accounts
	expectedAmount := ra.GenesisInfo.GenesisTransferAmount()
	if expectedAmount.String() != gTransfer.Amount {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "amount mismatch")
	}

	// split the transfer to the genesis accounts
	for _, acc := range ra.GenesisInfo.GenesisAccounts.Accounts {
		// create a new packet for each account
		data := transfertypes.NewFungibleTokenPacketData(
			gTransfer.Denom,
			acc.Amount.String(),
			gTransfer.Sender,
			acc.Address,
			"",
		)

		// mint and send tokens to the account
		// No event emitted, as we called the transfer keeper directly (vs the transfer middleware stack)
		err := w.transferKeeper.OnRecvPacket(ctx, packet, data)
		if err != nil {
			return errorsmod.Wrapf(err, "on receive packet: %s", acc.Address)
		}
	}

	return nil
}
