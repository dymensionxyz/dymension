package types_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/x/agent/types"
)

func TestAttestedTransferBytes(t *testing.T) {
	amount := math.NewInt(42)

	// empty invoice ref keeps the legacy layout byte-for-byte
	require.Equal(t, []byte("dym1r\x00adym\x0042\x00memo"), types.AttestedTransferBytes("dym1r", "adym", amount, nil, []byte("memo")))
	require.Equal(t, types.AttestedTransferBytes("dym1r", "adym", amount, nil, nil), types.AttestedTransferBytes("dym1r", "adym", amount, []byte{}, nil))

	ref := []byte("in\x00v")
	payload := types.AttestedTransferBytes("dym1r", "adym", amount, ref, []byte("me\x00mo"))
	require.Equal(t, []byte("dym1r\x00adym\x0042\x00\x00\x00\x00\x00\x00\x00\x00\x04in\x00vme\x00mo"), payload)
	require.Equal(t, payload, types.AttestedTransferBytes("dym1r", "adym", amount, ref, []byte("me\x00mo")))

	// the length prefix keeps the ref/memo boundary unambiguous, including
	// across embedded NULs
	splits := [][2][]byte{
		{[]byte("in"), []byte("\x00vme\x00mo")},
		{[]byte("in\x00vm"), []byte("e\x00mo")},
		{[]byte("in\x00vme\x00mo"), nil},
	}
	for _, split := range splits {
		require.NotEqual(t, payload, types.AttestedTransferBytes("dym1r", "adym", amount, split[0], split[1]))
	}
}

// TestTransferNonceInvoiceDomain pins that an invoice transfer's token cannot
// verify for a legacy transfer even when the payload bytes coincide (a legacy
// memo may start with the invoice length prefix).
func TestTransferNonceInvoiceDomain(t *testing.T) {
	amount := math.NewInt(42)
	ref := []byte("inv")
	invoiced := types.AttestedTransferBytes("dym1r", "adym", amount, ref, []byte("m"))
	legacy := types.AttestedTransferBytes("dym1r", "adym", amount, nil, append([]byte{0, 0, 0, 0, 0, 0, 0, 3}, "invm"...))
	require.Equal(t, invoiced, legacy)

	require.NotEqual(t, types.TransferNonce("a1", invoiced, 7, true), types.TransferNonce("a1", legacy, 7, false))
	require.NotEqual(t, types.TransferNonce("a1", invoiced, 7, true), types.ActionNonce("a1", invoiced, 7))
	require.NotEqual(t, types.TransferNonce("a1", invoiced, 7, true), types.ValidationNonce("a1", invoiced, 7))
	require.Equal(t, types.TransferNonce("a1", invoiced, 7, true), types.TransferNonce("a1", invoiced, 7, true))
}
