package types

import (
	hypercoretypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/01_interchain_security/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func (u *MsgIndicateProgress) ValidateBasic() error {
	if err := u.Payload.ValidateBasic(); err != nil {
		return err
	}

	_, err := u.ParseMetadata()
	if err != nil {
		return err
	}

	return nil
}

func (u *MsgIndicateProgress) ParseMetadata() (hypercoretypes.MessageIdMultisigRawMetadata, error) {
	if u.Metadata == nil {
		return hypercoretypes.MessageIdMultisigRawMetadata{}, gerrc.ErrInvalidArgument.Wrapf("metadata")
	}
	return hypercoretypes.NewMessageIdMultisigRawMetadata(u.Metadata)
}

func (u *MsgIndicateProgress) MustGetMetadata() hypercoretypes.MessageIdMultisigRawMetadata {
	metadata, err := u.ParseMetadata()
	if err != nil {
		panic(err)
	}
	return metadata
}
