package types

import (
	"errors"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var (
	_ sdk.Msg = &MsgUpdateRollappInformation{}
	_ sdk.Msg = &MsgForceGenesisInfoChange{}
)

/* ----------------------- MsgUpdateRollappInformation ---------------------- */
func NewMsgUpdateRollappInformation(
	creator,
	rollappId,
	initSequencer string,
	minSeqBond *sdk.Coin,
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

	// TODO: validate min seq bond (https://github.com/dymensionxyz/dymension/issues/1703)

	if msg.Metadata != nil {
		if err := msg.Metadata.Validate(); err != nil {
			return errors.Join(ErrInvalidMetadata, err)
		}
	}

	if msg.GenesisInfo != nil {
		if err := msg.GenesisInfo.ValidateBasic(); err != nil {
			return errors.Join(ErrInvalidGenesisInfo, err)
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
	if err := m.NewGenesisInfo.ValidateBasic(); err != nil {
		return errorsmod.Wrapf(
			errors.Join(gerrc.ErrInvalidArgument, err),
			"invalid genesis info",
		)
	}

	if !m.NewGenesisInfo.Launchable() {
		return errorsmod.Wrapf(
			errors.Join(gerrc.ErrInvalidArgument, fmt.Errorf("missing fields in genesis info")),
			"invalid genesis info",
		)
	}

	return nil
}
