package uhyp

import (
	hyperutil "github.com/bcp-innovations/hyperlane-cosmos/util"
)

func MustDecodeHexAddress(s string) hyperutil.HexAddress {
	addr, err := hyperutil.DecodeHexAddress(s)
	if err != nil {
		panic(err)
	}
	return addr
}
