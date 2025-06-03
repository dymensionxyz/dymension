package types

import (
	"github.com/bcp-innovations/hyperlane-cosmos/util"
	"github.com/cosmos/gogoproto/proto"
	gethcrypto "github.com/ethereum/go-ethereum/crypto"
)

func (id *WithdrawalID) DecodeMailboxId() (util.HexAddress, error) {
	return util.DecodeHexAddress(id.MailboxId)
}

func (id *WithdrawalID) DecodeMessageId() (util.HexAddress, error) {
	return util.DecodeHexAddress(id.MessageId)
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
