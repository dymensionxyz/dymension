package types

import (
	"errors"
	"fmt"
	"slices"

	errorsmod "cosmossdk.io/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uibc"
)

// HubRecipient is the address of `x/rollapp` module's account on the rollapp chain.
const HubRecipient = "dym1mk7pw34ypusacm29m92zshgxee3yreums8avur"

type GenesisBridgeValidator struct {
	rollapp   GenesisBridgeData // what the rollapp sent over IBC
	hub       GenesisInfo       // what the rollapp thinks is correct
	channelID string            // can use "channel-0" in simulation
	rollappID string            // the actual rollapp ID
}

func NewGenesisBridgeValidator(
	rollappGenesis GenesisBridgeData,
	hubGenesis GenesisInfo,
	rollappChannelID string,
	rollappID string,
) *GenesisBridgeValidator {
	return &GenesisBridgeValidator{
		rollapp:   rollappGenesis,
		hub:       hubGenesis,
		channelID: rollappChannelID,
		rollappID: rollappID,
	}
}

type ValidationResult struct {
	NativeDenomTrace  transfertypes.DenomTrace                // the ibc denom trace of the native rollapp denom
	NativeDenom       banktypes.Metadata                      // the metadata of the native rollapp denom
	GenesisAccPackets []transfertypes.FungibleTokenPacketData // packets holding native tokens dedicated for genesis accounts
}

func (v *GenesisBridgeValidator) Validate() (*ValidationResult, error) {
	if err := v.rollapp.ValidateBasic(); err != nil {
		return nil, errorsmod.Wrap(err, "validate basic genesis bridge data")
	}

	if err := validateAgainstHub(v.rollapp.GenesisInfo, v.hub); err != nil {
		return nil, errorsmod.Wrap(err, "validate against rollapp")
	}

	trace, denomMeta, err := v.validateNativeDenom()
	if err != nil {
		return nil, errorsmod.Wrap(err, "validate native denom")
	}

	genesisAccPackets, err := v.validateGenesisTransfer()
	if err != nil {
		return nil, errorsmod.Wrap(err, "validate genesis transfer")
	}

	return &ValidationResult{
		NativeDenomTrace:  trace,
		NativeDenom:       denomMeta,
		GenesisAccPackets: genesisAccPackets,
	}, nil
}

func validateAgainstHub(packet GenesisBridgeInfo, hub GenesisInfo) error {
	if packet.GenesisChecksum != hub.GenesisChecksum {
		return fmt.Errorf("genesis checksum mismatch: expected: %v, got: %v", hub.GenesisChecksum, packet.GenesisChecksum)
	}

	if packet.Bech32Prefix != hub.Bech32Prefix {
		return fmt.Errorf("bech32 prefix mismatch: expected: %v, got: %v", hub.Bech32Prefix, packet.Bech32Prefix)
	}

	if packet.NativeDenom != hub.NativeDenom {
		return fmt.Errorf("native denom mismatch: expected: %v, got: %v", hub.NativeDenom, packet.NativeDenom)
	}

	if !packet.InitialSupply.Equal(hub.InitialSupply) {
		return fmt.Errorf("initial supply mismatch: expected: %v, got: %v", hub.InitialSupply, packet.InitialSupply)
	}

	err := compareGenesisAccounts(hub.Accounts(), packet.GenesisAccounts)
	if err != nil {
		return errorsmod.Wrap(err, "genesis accounts mismatch")
	}
	return nil
}

func compareGenesisAccounts(raCommitted []GenesisAccount, gbData []GenesisAccount) error {
	if len(raCommitted) != len(gbData) {
		return fmt.Errorf("genesis accounts length mismatch: expected %d, got %d", len(raCommitted), len(gbData))
	}

	for _, acc := range raCommitted {
		found := slices.ContainsFunc(gbData, func(dataAcc GenesisAccount) bool {
			return dataAcc.Address == acc.Address && dataAcc.Amount.Equal(acc.Amount)
		})

		if !found {
			return fmt.Errorf("genesis account mismatch: account %s with amount %v not found in data", acc.Address, acc.Amount)
		}
	}

	return nil
}

// validateNativeDenom extracts the IBC denom trace and metadata from the rollapp native denom.
// Resulting IBC denom is validated.
func (v *GenesisBridgeValidator) validateNativeDenom() (transfertypes.DenomTrace, banktypes.Metadata, error) {
	m := v.rollapp.NativeDenom
	trace := uibc.GetForeignDenomTrace(v.channelID, m.Base)

	// Change the base to the ibc denom, and add an alias to the original
	m.Base = trace.IBCDenom()
	m.Description = fmt.Sprintf("auto-generated ibc denom for rollapp: base: %s: rollapp: %s", m.GetBase(), v.rollappID)
	for i, u := range m.DenomUnits {
		if u.Exponent == 0 {
			m.DenomUnits[i].Aliases = append(m.DenomUnits[i].Aliases, u.Denom)
			m.DenomUnits[i].Denom = m.GetBase()
		}
	}

	if err := m.Validate(); err != nil {
		return transfertypes.DenomTrace{}, banktypes.Metadata{}, errorsmod.Wrap(errors.Join(gerrc.ErrInvalidArgument, err), "metadata validate")
	}
	return trace, m, nil
}

// validateGenesisTransfer validates the genesis transfer and prepares the packets for each genesis account.
func (v *GenesisBridgeValidator) validateGenesisTransfer() ([]transfertypes.FungibleTokenPacketData, error) {
	gTransfer := v.rollapp.GenesisTransfer
	requiresTransfer := v.hub.RequiresTransfer()

	// required but not present
	if requiresTransfer && gTransfer == nil {
		return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "genesis transfer required")
	}
	// not required but present
	if !requiresTransfer && gTransfer != nil {
		return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "genesis transfer not expected")
	}
	if gTransfer == nil {
		return nil, nil
	}

	// validate the receiver
	if gTransfer.Receiver != HubRecipient {
		return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "receiver mismatch")
	}

	// validate that the transfer amount matches the expected amount, which is the sum of all genesis accounts
	expectedAmount := v.hub.GenesisTransferAmount()
	if expectedAmount.String() != gTransfer.Amount {
		return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "amount mismatch")
	}

	var ret []transfertypes.FungibleTokenPacketData
	for _, acc := range v.hub.Accounts() {
		// create a new packet for each account
		data := transfertypes.NewFungibleTokenPacketData(
			gTransfer.Denom,
			acc.Amount.String(),
			gTransfer.Sender,
			acc.Address,
			"",
		)
		ret = append(ret, data)
	}
	return ret, nil
}
