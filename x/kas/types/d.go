package types

import "github.com/bcp-innovations/hyperlane-cosmos/util"

func (id *WithdrawalID) DecodeMailboxId() (util.HexAddress, error) {
	return util.DecodeHexAddress(id.MailboxId)
}

func (id *WithdrawalID) DecodeMessageId() (util.HexAddress, error) {
	return util.DecodeHexAddress(id.MessageId)
}
