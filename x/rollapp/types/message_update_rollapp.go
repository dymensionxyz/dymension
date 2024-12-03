package types

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/migrations/legacytx"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

const TypeMsgUpdateRollappInformation = "update_rollapp"

var (
	_ sdk.Msg            = &MsgUpdateRollappInformation{}
	_ sdk.Msg            = &MsgForceGenesisInfoChange{}
	_ legacytx.LegacyMsg = &MsgUpdateRollappInformation{}
)

/* ----------------------- MsgUpdateRollappInformation ---------------------- */
func NewMsgUpdateRollappInformation(
	creator,
	rollappId,
	initSequencer string,
	minSeqBond sdk.Coin,
	metadata *RollappMetadata,
	genesisInfo *GenesisInfo,
) *MsgUpdateRollappInformation {
	return &MsgUpdateRollappInformation{
		Owner:            creator,
		RollappId:        rollappId,
		InitialSequencer: initSequencer,
		MinSequencerBond: minSeqBond,
		Metadata:         metadata,
		GenesisInfo:      genesisInfo,
	}
}

func (msg *MsgUpdateRollappInformation) Route() string {
	return RouterKey
}

func (msg *MsgUpdateRollappInformation) Type() string {
	return TypeMsgUpdateRollappInformation
}

func (msg *MsgUpdateRollappInformation) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpdateRollappInformation) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateRollappInformation) ValidateBasic() error {
	_, err := sdk.AccAddressFromBech32(msg.Owner)
	if err != nil {
		return errors.Join(ErrInvalidCreatorAddress, err)
	}

	if msg.InitialSequencer != "" && msg.InitialSequencer != "*" {
		_, err := sdk.AccAddressFromBech32(msg.InitialSequencer)
		if err != nil {
			return errors.Join(ErrInvalidInitialSequencer, err)
		}
	}

	if msg.GenesisInfo != nil {
		if err := msg.GenesisInfo.Validate(); err != nil {
			return err
		}
	}

	if msg.Metadata != nil {
		if err := msg.Metadata.Validate(); err != nil {
			return errors.Join(ErrInvalidMetadata, err)
		}
	}

	return nil
}

func (msg *MsgUpdateRollappInformation) UpdatingImmutableValues() bool {
	return msg.InitialSequencer != "" || IsUpdateMinSeqBond(msg.MinSequencerBond)
}

func (msg *MsgUpdateRollappInformation) UpdatingGenesisInfo() bool {
	return msg.GenesisInfo != nil
}

/* ------------------------ MsgForceGenesisInfoChange ----------------------- */
// ValidateBasic performs basic validation for the MsgForceGenesisInfoChange.
func (m *MsgForceGenesisInfoChange) ValidateBasic() error {
	// Validate authority address
	_, err := sdk.AccAddressFromBech32(m.Authority)
	if err != nil {
		return errorsmod.Wrapf(
			errors.Join(gerrc.ErrInvalidArgument, err),
			"authority is not a valid bech32 address: %s", m.Authority,
		)
	}

	// Validate rollapp ID
	if m.RollappId == "" {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "rollapp_id cannot be empty")
	}

	// Validate new genesis info
	if err := m.NewGenesisInfo.Validate(); err != nil {
		return errorsmod.Wrapf(
			errors.Join(gerrc.ErrInvalidArgument, err),
			"invalid genesis info",
		)
	}

	if !m.NewGenesisInfo.AllSet() {
		return errorsmod.Wrapf(
			errors.Join(gerrc.ErrInvalidArgument, fmt.Errorf("missing fields in genesis info")),
			"invalid genesis info",
		)
	}

	return nil
}

// GetSigners returns the expected signers for a MsgForceGenesisInfoChange.
func (m *MsgForceGenesisInfoChange) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}
