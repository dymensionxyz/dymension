package uaddr_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/utils/uaddr"
)

func TestBech32ToAddr(t *testing.T) {
	testCases := []struct {
		validPrefix string
	}{
		{validPrefix: "cosmosvaloper"},
		{validPrefix: "dymval"},
		{validPrefix: "cosmos"},
		{validPrefix: "dym"},
	}

	for _, tc := range testCases {
		t.Run(tc.validPrefix, func(t *testing.T) {
			// generate an address with a random prefix
			expectedBytes := ed25519.GenPrivKey().PubKey().Address().Bytes()
			randomPrefixedAddr, err := bech32.ConvertAndEncode(tc.validPrefix, expectedBytes)
			require.NoError(t, err)

			// convert the random-prefixed bech32 to the current val oper address
			valOperAddr, err := uaddr.FromBech32[sdk.ValAddress](randomPrefixedAddr)
			require.NoError(t, err)

			// check results
			// verify format
			err = sdk.VerifyAddressFormat(valOperAddr)
			require.NoError(t, err)

			// cast valOperAddr to string and verify the prefix and bytes
			actualPrefix, actualBytes, err := bech32.DecodeAndConvert(valOperAddr.String())
			require.NoError(t, err)
			expectedPrefix := sdk.GetConfig().GetBech32ValidatorAddrPrefix()
			require.Equal(t, expectedPrefix, actualPrefix)
			require.Equal(t, expectedBytes, actualBytes)
		})
	}
}
