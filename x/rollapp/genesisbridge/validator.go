package genesisbridge

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	transfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uibc"
)

type validator struct {
	rollapp   GenesisBridgeData // what the rollapp sent over IBC
	hub       types.GenesisInfo // what the hub thinks is correct
	channelID string            // can use "channel-0" in simulation
	rollappID string            // the actual rollapp ID
}

type actionItems struct {
	trace      transfertypes.DenomTrace
	bankMeta   banktypes.Metadata
	fungiDatas []transfertypes.FungibleTokenPacketData
}

func (v *validator) validateAndGetActionItems() (*actionItems, error) {
	if err := v.rollapp.ValidateBasic(); err != nil {
		return nil, errorsmod.Wrap(err, "validate basic genesis bridge data")
	}

	if err := v.validateAgainstHub(v.rollapp.GenesisInfo, v.hub); err != nil {
		return nil, errorsmod.Wrap(err, "validate against hub")
	}

	ret := &actionItems{}
	trace, bankMeta, err := v.getBankStuff()
	if err != nil {
		return nil, errorsmod.Wrap(err, "get bank stuff")
	}
	ret.trace = *trace
	ret.bankMeta = *bankMeta

	fungiDatas, err := v.getFungiData()
	if err != nil {
		return nil, errorsmod.Wrap(err, "get fungi data")
	}
	ret.fungiDatas = fungiDatas
	return ret, nil
}

func (v *validator) validateAgainstHub(packet GenesisBridgeInfo, hub types.GenesisInfo) error {
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

	err := compareGenesisAccounts(hub.GenesisAccounts, packet.GenesisAccounts)
	if err != nil {
		return errorsmod.Wrap(err, "genesis accounts mismatch")
	}
	return nil
}

func compareGenesisAccounts(raCommitted *types.GenesisAccounts, gbData []types.GenesisAccount) error {
	if raCommitted == nil {
		if len(gbData) == 0 {
			return nil
		}
		return fmt.Errorf("genesis accounts length mismatch: expected 0, got %d", len(gbData))
	}

	if len(raCommitted.Accounts) != len(gbData) {
		return fmt.Errorf("genesis accounts length mismatch: expected %d, got %d", len(raCommitted.Accounts), len(gbData))
	}

	for _, acc := range raCommitted.Accounts {
		found := slices.ContainsFunc(gbData, func(dataAcc types.GenesisAccount) bool {
			return dataAcc.Address == acc.Address && dataAcc.Amount.Equal(acc.Amount)
		})

		if !found {
			return fmt.Errorf("genesis account mismatch: account %s with amount %v not found in data", acc.Address, acc.Amount)
		}
	}

	return nil
}

func (v *validator) getBankStuff() (*transfertypes.DenomTrace, *banktypes.Metadata, error) {
	var m banktypes.Metadata
	m = v.rollapp.NativeDenom
	trace := uibc.GetForeignDenomTrace(v.channelID, m.Base)

	// Change the base to the ibc denom, and add an alias to the original
	m.Base = trace.IBCDenom()
	m.Description = fmt.Sprintf("auto-generated ibc denom for hub: base: %s: hub: %s", m.GetBase(), v.rollappID)
	for i, u := range m.DenomUnits {
		if u.Exponent == 0 {
			m.DenomUnits[i].Aliases = append(m.DenomUnits[i].Aliases, u.Denom)
			m.DenomUnits[i].Denom = m.GetBase()
		}
	}

	if err := m.Validate(); err != nil {
		return nil, nil, errorsmod.Wrap(errors.Join(gerrc.ErrInvalidArgument, err), "metadata validate")
	}
	return &trace, &m, nil
}

func (v *validator) getFungiData() ([]transfertypes.FungibleTokenPacketData, error) {
	gTransfer := v.rollapp.GenesisTransfer
	required := v.hub.GenesisAccounts != nil
	// required but not present
	if required && gTransfer == nil {
		return nil, errorsmod.Wrap(gerrc.ErrFailedPrecondition, "genesis transfer required")
	}
	// not required but present
	if !required && gTransfer != nil {
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
	for _, acc := range v.hub.GenesisAccounts.Accounts {
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
