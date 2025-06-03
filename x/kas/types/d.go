package types

import (
	"bytes"

	"github.com/bcp-innovations/hyperlane-cosmos/util"
	"github.com/cosmos/gogoproto/proto"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	gethcrypto "github.com/ethereum/go-ethereum/crypto"
)

func (msg *ProgressIndication) ValidateBasic() error {
	if msg == nil {
		return gerrc.ErrInvalidArgument.Wrapf("progress indication is nil")
	}

	if err := msg.OldOutpoint.ValidateBasic(); err != nil {
		return err
	}

	if err := msg.NewOutpoint.ValidateBasic(); err != nil {
		return err
	}

	for _, sig := range msg.ProcessedWithdrawals {
		if err := sig.ValidateBasic(); err != nil {
			return err
		}
	}

	return nil
}

func (msg *TransactionOutpoint) ValidateBasic() error {
	if msg == nil {
		return gerrc.ErrInvalidArgument.Wrapf("transaction outpoint is nil")
	}

	if len(msg.TransactionId) != 32 {
		return gerrc.ErrInvalidArgument.Wrapf("transaction id")
	}

	return nil
}

func (o *TransactionOutpoint) Equal(other *TransactionOutpoint) bool {
	return bytes.Equal(o.TransactionId, other.TransactionId) && o.Index == other.Index
}

func (id *WithdrawalID) ValidateBasic() error {
	if id == nil {
		return gerrc.ErrInvalidArgument.Wrapf("withdrawal id is nil")
	}

	if _, err := id.DecodeMailboxId(); err != nil {
		return err
	}

	if _, err := id.DecodeMessageId(); err != nil {
		return err
	}
	return nil
}

func (id *WithdrawalID) DecodeMailboxId() (util.HexAddress, error) {
	return util.DecodeHexAddress(id.MailboxId)
}

func (id *WithdrawalID) DecodeMessageId() (util.HexAddress, error) {
	return util.DecodeHexAddress(id.MessageId)
}

func (id *WithdrawalID) MustMailboxId() util.HexAddress {
	ret, _ := util.DecodeHexAddress(id.MailboxId)
	return ret
}

func (id *WithdrawalID) MustMessageId() util.HexAddress {
	ret, _ := util.DecodeHexAddress(id.MessageId)
	return ret
}

// returns what should be signed by validators
// see https://github.com/dymensionxyz/hyperlane-cosmos/blob/fb914a5ba702f70a428a475968b886891cb1ad77/x/core/01_interchain_security/types/merkle_root_multisig.go#L163-L173
func (u *ProgressIndication) SignBytes() ([32]byte, error) {
	bz, err := proto.Marshal(u) // TODO: check. Gogoproto should be fine
	if err != nil {
		return [32]byte{}, err
	}

	kec := gethcrypto.Keccak256(bz)
	return util.GetEthSigningHash(kec), nil
}
