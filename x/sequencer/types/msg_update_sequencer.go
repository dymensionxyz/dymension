package types

import (
	errorsmod "cosmossdk.io/errors"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/dymensionxyz/dymension/v3/x/rollapp/types"
)

var (
	_ sdk.Msg                            = &MsgUpdateSequencerInformation{}
	_ codectypes.UnpackInterfacesMessage = (*MsgUpdateSequencerInformation)(nil)
)

func NewMsgUpdateSequencerInformation(creator string, rollappId string, metadata *SequencerMetadata) (*MsgUpdateSequencerInformation, error) {
	if metadata == nil {
		return nil, ErrInvalidRequest
	}
	return &MsgUpdateSequencerInformation{
		Creator:   creator,
		RollappId: rollappId,
		Metadata:  *metadata,
	}, nil
}

func (msg *MsgUpdateSequencerInformation) Route() string {
	return RouterKey
}

func (msg *MsgUpdateSequencerInformation) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgUpdateSequencerInformation) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgUpdateSequencerInformation) ValidateBasic() error {
	if err := msg.Metadata.Validate(); err != nil {
		return errorsmod.Wrap(ErrInvalidMetadata, err.Error())
	}

	return nil
}

func (msg *MsgUpdateSequencerInformation) VMSpecificValidate(vmType types.Rollapp_VMType) error {
	if vmType == types.Rollapp_EVM {
		if err := validateURLs(msg.Metadata.EvmRpcs); err != nil {
			return errorsmod.Wrap(err, "invalid evm rpcs URLs")
		}
	}
	return nil
}

func (msg *MsgUpdateSequencerInformation) UnpackInterfaces(codectypes.AnyUnpacker) error { return nil }
