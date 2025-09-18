package types

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

var _ sdk.Msg = &MsgCreateRollapp{}

func NewMsgCreateRollapp(
	creator,
	rollappId,
	initSequencer string,
	minSequencerBond sdk.Coin,
	alias string,
	vmType Rollapp_VMType,
	metadata *RollappMetadata,
	genesisInfo *GenesisInfo,
	feeDenom string,
) *MsgCreateRollapp {
	return &MsgCreateRollapp{
		Creator:          creator,
		RollappId:        rollappId,
		InitialSequencer: initSequencer,
		MinSequencerBond: minSequencerBond,
		Alias:            alias,
		VmType:           vmType,
		Metadata:         metadata,
		GenesisInfo:      genesisInfo,
		FeeDenom:         feeDenom,
	}
}

func (msg *MsgCreateRollapp) GetRollapp() Rollapp {
	genInfo := GenesisInfo{}
	if msg.GenesisInfo != nil {
		genInfo = *msg.GenesisInfo
		// hotfix: if supply is zero, override the denom metadata with empty
		if genInfo.InitialSupply.IsZero() {
			genInfo.NativeDenom = DenomMetadata{}
		}
	}
	return NewRollapp(
		msg.Creator,
		msg.RollappId,
		msg.InitialSequencer,
		msg.MinSequencerBond,
		msg.VmType,
		msg.Metadata,
		genInfo,
	)
}

func (msg *MsgCreateRollapp) ValidateBasic() error {
	if len(msg.Alias) == 0 {
		return ErrInvalidAlias
	}

	if msg.FeeDenom != "" && sdk.ValidateDenom(msg.FeeDenom) != nil {
		return errorsmod.Wrap(gerrc.ErrInvalidArgument, "invalid fee denom")
	}

	rollapp := msg.GetRollapp()
	if err := rollapp.ValidateBasic(); err != nil {
		return err
	}

	return nil
}
