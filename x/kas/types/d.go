package types

import (
	"github.com/bcp-innovations/hyperlane-cosmos/util"
	"github.com/cosmos/gogoproto/proto"
)

func (id *WithdrawalID) DecodeMailboxId() (util.HexAddress, error) {
	return util.DecodeHexAddress(id.MailboxId)
}

func (id *WithdrawalID) DecodeMessageId() (util.HexAddress, error) {
	return util.DecodeHexAddress(id.MessageId)
}

// returns what should be signed by validators
func (u *ProgressIndication) SignBytes() ([]byte, error) {
	return proto.Marshal(u)
}
