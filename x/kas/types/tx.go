package types

import "github.com/dymensionxyz/gerr-cosmos/gerrc"

func (msg *MsgIndicateProgress) ValidateBasic() error {
	if msg.Metadata == nil {
		return gerrc.ErrInvalidArgument.Wrapf("metadata")
	}

	if msg.Payload == nil {
		return gerrc.ErrInvalidArgument.Wrapf("payload")
	}

	return msg.Payload.ValidateBasic()
}
