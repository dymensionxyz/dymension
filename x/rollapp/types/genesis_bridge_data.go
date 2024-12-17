package types

import (
	"fmt"

	"cosmossdk.io/errors"
	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uibc"
	"github.com/dymensionxyz/sdk-utils/utils/uslice"
)

// ValidateBasic performs basic validation checks on the GenesisBridgeData.
func (d GenesisBridgeData) ValidateBasic() error {
	if err := d.GenesisInfo.ValidateBasic(); err != nil {
		return errors.Wrap(err, "invalid genesis info")
	}

	if d.GenesisInfo.NativeDenom.IsSet() {
		if err := d.NativeDenom.Validate(); err != nil {
			return errors.Wrap(err, "invalid metadata")
		}

		if d.NativeDenom.Base != d.GenesisInfo.NativeDenom.Base {
			return fmt.Errorf("metadata denom does not match genesis info denom")
		}

		// validate the decimals of the display denom
		valid := false
		for _, unit := range d.NativeDenom.DenomUnits {
			if unit.Denom == d.GenesisInfo.NativeDenom.Display {
				if unit.Exponent == d.GenesisInfo.NativeDenom.Exponent {
					valid = true
					break
				}
			}
		}
		if !valid {
			return fmt.Errorf("denom metadata does not contain display unit with the correct exponent")
		}
	}

	if d.GenesisTransfer != nil {
		if err := d.GenesisTransfer.ValidateBasic(); err != nil {
			return errors.Wrap(err, "invalid genesis transfer")
		}

		if d.GenesisInfo.NativeDenom.Base != d.GenesisTransfer.Denom {
			return errorsmod.Wrap(gerrc.ErrFailedPrecondition, "denom mismatch")
		}
	}

	return nil
}

// IBCDenom extracts the IBC denom trace and metadata from the rollapp native denom.
func (d GenesisBridgeData) IBCDenom(rollappID, channelID string) (transfertypes.DenomTrace, banktypes.Metadata, error) {
	m := d.NativeDenom
	trace := uibc.GetForeignDenomTrace(channelID, m.Base)

	// Change the base to the ibc denom, and add an alias to the original
	m.Base = trace.IBCDenom()
	m.Description = fmt.Sprintf("auto-generated ibc denom for rollapp: base: %s: rollapp: %s", m.GetBase(), rollappID)
	for i, u := range m.DenomUnits {
		if u.Exponent == 0 {
			m.DenomUnits[i].Aliases = append(m.DenomUnits[i].Aliases, u.Denom)
			m.DenomUnits[i].Denom = m.Base
		}
	}

	if err := m.Validate(); err != nil {
		return transfertypes.DenomTrace{}, banktypes.Metadata{}, fmt.Errorf("validate IBC denom metadata: %w", err)
	}
	return trace, m, nil
}

// GenesisAccPackets creates a new packet for each genesis account.
func (d *GenesisBridgeData) GenesisAccPackets() []transfertypes.FungibleTokenPacketData {
	return uslice.Map(d.GenesisInfo.Accounts(), func(acc GenesisAccount) transfertypes.FungibleTokenPacketData {
		return transfertypes.NewFungibleTokenPacketData(
			d.GenesisTransfer.Denom,
			acc.Amount.String(),
			d.GenesisTransfer.Sender,
			acc.Address,
			"",
		)
	})
}

// Handling should be based on length and contents, not nil status
func (i GenesisBridgeInfo) Accounts() []GenesisAccount {
	if i.GenesisAccounts == nil {
		return nil
	}
	return i.GenesisAccounts
}

func (i GenesisBridgeInfo) RequiresTransfer() bool {
	return 0 < len(i.Accounts())
}

func (i GenesisBridgeInfo) GenesisTransferAmount() math.Int {
	total := math.ZeroInt()
	for _, a := range i.Accounts() {
		total = total.Add(a.Amount)
	}
	return total
}

// converts to a native type and validates that
func (i GenesisBridgeInfo) ValidateBasic() error {
	// wrap the genesis info in a GenesisInfo struct, to reuse the validation logic
	raGenesisInfo := GenesisInfo{
		GenesisChecksum: i.GenesisChecksum,
		Bech32Prefix:    i.Bech32Prefix,
		NativeDenom:     i.NativeDenom,
		InitialSupply:   i.InitialSupply,
		GenesisAccounts: &GenesisAccounts{Accounts: i.GenesisAccounts},
	}

	if !raGenesisInfo.Launchable() {
		return fmt.Errorf("missing fields in genesis bridge info")
	}

	return raGenesisInfo.ValidateBasic()
}
