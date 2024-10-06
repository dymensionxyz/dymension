package genesisbridge

import (
	"encoding/json"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	gerrc "github.com/dymensionxyz/gerr-cosmos/gerrc"
)

const (
	// hubRecipient is the address of `x/rollapp` module's account on the hub chain.
	hubRecipient = "dym1mk7pw34ypusacm29m92zshgxee3yreums8avur"

	memoNamespaceKey = "genesis_transfer"
)

type GenesisTransferMemo struct {
	GenesisTransferData GenesisTransferData `json:"genesis_transfer"`
}

type GenesisTransferData struct {
	GenesisAccounts []types.GenesisAccount `json:"genesis_accounts"`
}

func (memo GenesisTransferMemo) MustMarshal() []byte {
	memoBytes, err := json.Marshal(memo)
	if err != nil {
		panic(err)
	}
	return memoBytes
}

func (memo GenesisTransferMemo) String() string {
	return string(memo.MustMarshal())
}

func (g GenesisTransferMemo) ValidateBasic() error {
	for _, acc := range g.GenesisTransferData.GenesisAccounts {
		if err := acc.ValidateBasic(); err != nil {
			return errorsmod.Wrap(err, "genesis account")
		}
	}
	return nil
}

// MustString returns a human-readable json string - intended for tests.
func (g GenesisTransferMemo) MustString() string {
	bz, err := json.MarshalIndent(g, "", "\t")
	if err != nil {
		panic(err)
	}
	return string(bz)
}

// HandleGenesisTransfer handles the genesis transfer packet, if present, and expected.
func (w IBCModule) handleGenesisTransfer(ctx sdk.Context, ra types.Rollapp, packet channeltypes.Packet, gTransfer *transfertypes.FungibleTokenPacketData) error {
	// check if required or expected
	required := len(ra.GenesisInfo.GenesisAccounts) > 0
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
	if gTransfer.Receiver != hubRecipient {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "receiver mismatch")
	}

	// extract genesis transfer memo if exists
	memo, err := getMemo(gTransfer.GetMemo())
	if err != nil {
		return errorsmod.Wrap(err, "get memo")
	}

	// validate the genesis transfer memo against the rollapp expected genesis accounts
	err = compareMemoWithRollapp(ra, memo)
	if err != nil {
		return errorsmod.Wrap(err, "compare memo with rollapp")
	}

	// validate the genesis transfer denom
	if ra.GenesisInfo.NativeDenom.Base != gTransfer.Denom {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "denom mismatch")
	}

	// validate that the transfer amount matches the expected amount, which is the sum of all genesis accounts
	expectedAmount := ra.GenesisInfo.GenesisTransferAmount()
	if expectedAmount.String() != gTransfer.Amount {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "amount mismatch")
	}

	// split the transfer to the genesis accounts
	for _, acc := range memo.GenesisTransferData.GenesisAccounts {
		// create a new packet for each account
		data := transfertypes.NewFungibleTokenPacketData(
			gTransfer.Sender,
			acc.Address,
			gTransfer.Denom,
			acc.Amount.String(),
			"",
		)

		// mint and send tokens to the account
		// No event emitted, as we called the transfer keeper directly (vs the transfer middleware stack)
		err = w.transferKeeper.OnRecvPacket(ctx, packet, data)
		if err != nil {
			return errorsmod.Wrapf(err, "on receive packet: %s", acc.Address)
		}
	}

	return nil

}

// FIXME: make it order insensitive
// validate the genesis transfer memo against the rollapp expected genesis accounts
func compareMemoWithRollapp(ra types.Rollapp, memo GenesisTransferMemo) error {
	gacc := ra.GenesisInfo.GenesisAccounts
	if len(gacc) != len(memo.GenesisTransferData.GenesisAccounts) {
		return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "genesis accounts length mismatch")
	}

	for i, acc := range gacc {
		if acc.Address != memo.GenesisTransferData.GenesisAccounts[i].Address {
			return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "genesis account address mismatch")
		}
		if !acc.Amount.Equal(memo.GenesisTransferData.GenesisAccounts[i].Amount) {
			return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "genesis account amount mismatch")
		}
	}

	return nil
}

func getMemo(rawMemo string) (GenesisTransferMemo, error) {
	if len(rawMemo) == 0 {
		return GenesisTransferMemo{}, gerrc.ErrNotFound
	}

	var m GenesisTransferMemo
	err := json.Unmarshal([]byte(rawMemo), &m)
	if err != nil {
		return GenesisTransferMemo{}, errorsmod.Wrap(err, "unmarshal memo")
	}

	err = m.ValidateBasic()
	if err != nil {
		return GenesisTransferMemo{}, errorsmod.Wrap(err, "validate basic")
	}

	return m, nil
}
