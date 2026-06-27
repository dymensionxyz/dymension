package types_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/app/params"
	"github.com/dymensionxyz/dymension/v3/x/agent/types"
	"github.com/dymensionxyz/dymension/v3/x/common/tee"
)

const sampleOwner = "dym1wg8p6j0pxpnsvhkwfu54ql62cnrumf0v634mft"

func init() {
	params.SetAddressPrefixes(sdk.GetConfig())
}

func validCertPEM(t *testing.T) string {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)
	tmpl := &x509.Certificate{SerialNumber: big.NewInt(1), Subject: pkix.Name{CommonName: "test"}}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	require.NoError(t, err)
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
}

func TestMsgRegisterAgent_ValidateBasic(t *testing.T) {
	cert := validCertPEM(t)
	good := tee.Policy{GcpRootCertPem: cert, PolicyValues: "{}", PolicyQuery: "data.x.allow", PolicyStructure: "package x"}

	for _, tc := range []struct {
		name    string
		msg     *types.MsgRegisterAgent
		wantErr bool
	}{
		{"valid", types.NewMsgRegisterAgent(sampleOwner, "a1", good), false},
		{"empty id", types.NewMsgRegisterAgent(sampleOwner, "", good), true},
		{"bad owner", types.NewMsgRegisterAgent("not-an-address", "a1", good), true},
		{"bad cert", types.NewMsgRegisterAgent(sampleOwner, "a1", tee.Policy{GcpRootCertPem: "nope", PolicyQuery: "q", PolicyStructure: "s"}), true},
		{"empty query", types.NewMsgRegisterAgent(sampleOwner, "a1", tee.Policy{GcpRootCertPem: cert, PolicyStructure: "s"}), true},
		{"empty structure", types.NewMsgRegisterAgent(sampleOwner, "a1", tee.Policy{GcpRootCertPem: cert, PolicyQuery: "q"}), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.msg.ValidateBasic()
			if tc.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMsgDeactivateAgent_ValidateBasic(t *testing.T) {
	require.NoError(t, types.NewMsgDeactivateAgent(sampleOwner, "a1").ValidateBasic())
	require.Error(t, types.NewMsgDeactivateAgent(sampleOwner, "").ValidateBasic())
	require.Error(t, types.NewMsgDeactivateAgent("bad", "a1").ValidateBasic())
}
