package types

import (
	"bytes"
	"encoding/binary"

	"github.com/bcp-innovations/hyperlane-cosmos/util"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	gethcrypto "github.com/ethereum/go-ethereum/crypto"
)

func (u *ProgressIndication) ValidateBasic() error {
	if u == nil {
		return gerrc.ErrInvalidArgument.Wrapf("progress indication is nil")
	}

	if err := u.OldOutpoint.ValidateBasic(); err != nil {
		return err
	}

	if err := u.NewOutpoint.ValidateBasic(); err != nil {
		return err
	}

	for _, sig := range u.ProcessedWithdrawals {
		if err := sig.ValidateBasic(); err != nil {
			return err
		}
	}

	if _, err := u.SignBytes(); err != nil {
		return err
	}

	return nil
}

func (o *TransactionOutpoint) ValidateBasic() error {
	if o == nil {
		return gerrc.ErrInvalidArgument.Wrapf("transaction outpoint is nil")
	}

	if len(o.TransactionId) != 32 {
		return gerrc.ErrInvalidArgument.Wrapf("transaction id length")
	}

	return nil
}

func (o *TransactionOutpoint) Equal(other *TransactionOutpoint) bool {
	return bytes.Equal(o.TransactionId, other.TransactionId) && o.Index == other.Index
}

func (o *TransactionOutpoint) SignBytes() []byte {
	ret := make([]byte, 32)
	copy(ret, o.TransactionId)
	ix := make([]byte, 4)
	binary.BigEndian.PutUint32(ix, o.Index)
	return append(ret, ix...)
}

func (i *WithdrawalID) ValidateBasic() error {
	if i == nil {
		return gerrc.ErrInvalidArgument.Wrapf("withdrawal id is nil")
	}

	if _, err := i.DecodeMessageId(); err != nil {
		return err
	}
	return nil
}

func (i *WithdrawalID) DecodeMessageId() (util.HexAddress, error) {
	return util.DecodeHexAddress(i.MessageId)
}

func (i *WithdrawalID) MustMessageId() util.HexAddress {
	ret, _ := util.DecodeHexAddress(i.MessageId)
	return ret
}

func (i *WithdrawalID) SignBytes() []byte {
	// it's already in hex so we just take the value directly
	return []byte(i.MessageId)
}

// returns what should be signed by validators
// see https://github.com/dymensionxyz/hyperlane-cosmos/blob/fb914a5ba702f70a428a475968b886891cb1ad77/x/core/01_interchain_security/types/merkle_root_multisig.go#L163-L173
func (u *ProgressIndication) SignBytes() ([32]byte, error) {
	old := u.OldOutpoint.SignBytes()
	new := u.NewOutpoint.SignBytes()
	bz := append(old, new...)
	for _, w := range u.ProcessedWithdrawals {
		bz = append(bz, w.SignBytes()...)
	}
	kec := gethcrypto.Keccak256(bz)
	return util.GetEthSigningHash(kec), nil
}

func (u *ProgressIndication) MustGetSignBytes() [32]byte {
	ret, err := u.SignBytes()
	if err != nil {
		panic(err)
	}
	return ret
}
