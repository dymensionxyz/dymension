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

func (msg *ProgressIndication) ValidateBasic() error {
	if msg.OldOutpoint == nil {
		return gerrc.ErrInvalidArgument.Wrapf("old outpoint")
	}

	if msg.NewOutpoint == nil {
		return gerrc.ErrInvalidArgument.Wrapf("new outpoint")
	}

	return nil
}

func (msg *TransactionOutpoint) ValidateBasic() error {
	if len(msg.TransactionId) != 32 {
		return gerrc.ErrInvalidArgument.Wrapf("transaction id")
	}

	return nil
}
