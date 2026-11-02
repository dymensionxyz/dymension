package ledgercosmos

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

// fakeTransport records APDUs and replies with queued responses.
type fakeTransport struct {
	sent    [][]byte
	replies []reply
}

type reply struct {
	data   []byte
	status uint16
}

func (f *fakeTransport) Exchange(apdu []byte) ([]byte, uint16, error) {
	f.sent = append(f.sent, append([]byte(nil), apdu...))
	r := f.replies[0]
	f.replies = f.replies[1:]
	return r.data, r.status, nil
}

func (f *fakeTransport) Close() error { return nil }

var dymPath = []uint32{44, 60, 0, 0, 0}

// m/44'/60'/0'/0/0, little endian, first three levels hardened.
const dymPathHex = "2c000080" + "3c000080" + "00000080" + "00000000" + "00000000"

func TestSignSendsHRPInFirstChunk(t *testing.T) {
	msg := bytes.Repeat([]byte("a"), chunkSize+10)
	f := &fakeTransport{replies: []reply{{nil, swOK}, {nil, swOK}, {[]byte{0x30, 0x44}, swOK}}}

	sig, err := New(f, "dym").SignSECP256K1(dymPath, msg, 0)
	require.NoError(t, err)
	require.Equal(t, []byte{0x30, 0x44}, sig)
	require.Len(t, f.sent, 3)

	init := f.sent[0]
	require.Equal(t, []byte{cla, insSign, chunkInit, 0}, init[:4])
	require.Equal(t, dymPathHex+"03"+hex.EncodeToString([]byte("dym")), hex.EncodeToString(init[5:]))
	require.Equal(t, int(init[4]), len(init)-5)

	require.Equal(t, []byte{cla, insSign, chunkAdd, 0, chunkSize}, f.sent[1][:5])
	require.Equal(t, msg[:chunkSize], f.sent[1][5:])
	require.Equal(t, []byte{cla, insSign, chunkLast, 0, 10}, f.sent[2][:5])
	require.Equal(t, msg[chunkSize:], f.sent[2][5:])
}

func TestSignPassesTextualP2(t *testing.T) {
	f := &fakeTransport{replies: []reply{{nil, swOK}, {[]byte{1}, swOK}}}
	_, err := New(f, "dym").SignSECP256K1(dymPath, []byte("{}"), 1)
	require.NoError(t, err)
	require.Equal(t, byte(1), f.sent[0][3])
	require.Equal(t, byte(1), f.sent[1][3])
}

// A Ledger public key at m/44'/60'/0'/0/0 with its eth_secp256k1 and secp256k1 addresses.
var (
	testPub, _  = base64.StdEncoding.DecodeString("AtHB9dv54BVZVQXQVl4NHLnYf3llDc8ekpg5UCubnfBt")
	testEthAddr = "dym1f3t7qs0r3tfzvxrfszh34dnm8y0qwnzdg06cu3"
	testSecAddr = "dym1y5glh9s8vtf7dmx459kcp9fn2uu0g4d2lx5uda"
)

func addrReply(pub []byte, addr string) reply {
	return reply{append(append([]byte(nil), pub...), addr...), swOK}
}

func TestGetAddress(t *testing.T) {
	pub, addr := testPub, testEthAddr
	f := &fakeTransport{replies: []reply{addrReply(pub, addr), addrReply(pub, addr)}}
	app := New(f, "dym")

	gotPub, gotAddr, err := app.GetAddressPubKeySECP256K1(dymPath, "dym")
	require.NoError(t, err)
	require.Equal(t, pub, gotPub)
	require.Equal(t, addr, gotAddr)
	require.Equal(t, []byte{cla, insGetAddr, 1, 0}, f.sent[0][:4])
	require.Equal(t, "03"+hex.EncodeToString([]byte("dym"))+dymPathHex, hex.EncodeToString(f.sent[0][5:]))

	gotPub, err = app.GetPublicKeySECP256K1(dymPath)
	require.NoError(t, err)
	require.Equal(t, pub, gotPub)
	require.Equal(t, byte(0), f.sent[1][2], "public key lookups must not ask for confirmation")
}

func TestGetAddressRejectsMismatch(t *testing.T) {
	f := &fakeTransport{replies: []reply{addrReply(testPub, testSecAddr)}}
	_, _, err := New(f, "dym").GetAddressPubKeySECP256K1(dymPath, "dym")
	require.ErrorContains(t, err, "expected the eth_secp256k1 address "+testEthAddr)
}

func TestStatusError(t *testing.T) {
	f := &fakeTransport{replies: []reply{{nil, 0x698C}}}
	_, err := New(f, "dym").GetPublicKeySECP256K1(dymPath)
	var se *StatusError
	require.ErrorAs(t, err, &se)
	require.Equal(t, uint16(0x698C), se.Status)
	require.ErrorContains(t, err, "chain config not supported")
}

func TestInvalidPath(t *testing.T) {
	_, err := New(&fakeTransport{}, "dym").SignSECP256K1([]uint32{44, 60, 0}, []byte("{}"), 0)
	require.ErrorContains(t, err, "5 levels")
}
