// Package ledgercosmos is a client for the Ledger Cosmos app that signs for Dymension's
// Ethereum-style (eth_secp256k1) accounts.
//
// The Cosmos app has a chain configuration for coin type 60 and the "dym" HRP: it shows the
// Ethereum-style address (the one Keplr and MetaMask show) and signs the keccak256 hash of the
// amino JSON sign bytes, which is what ethsecp256k1.PubKey verifies. On coin type 60 the app
// requires the HRP in the first chunk of a sign request. ledger-cosmos-go does not send it, so
// the SDK Ledger client is rejected with 0x698C (chain config not supported). This client
// sends it; the transaction is shown on the device (no blind signing).
package ledgercosmos

import (
	"encoding/binary"
	"errors"
	"fmt"

	sdkledger "github.com/cosmos/cosmos-sdk/crypto/ledger"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
)

const (
	cla = 0x55

	insGetVersion = 0x00
	insSign       = 0x02
	insGetAddr    = 0x04

	chunkInit = 0x00
	chunkAdd  = 0x01
	chunkLast = 0x02

	chunkSize   = 250
	hardenCount = 3
	pubKeyLen   = 33

	swOK = 0x9000
)

// statusText describes the Cosmos app status words users are most likely to hit.
var statusText = map[uint16]string{
	0x6984: "data invalid (on coin type 60 the Cosmos app requires Expert mode: Settings > Expert mode)",
	0x6985: "rejected on the device",
	0x6986: "rejected on the device",
	0x698C: "chain config not supported by the Cosmos app for this coin type and HRP",
	0x6D00: "instruction not supported: open the Cosmos app on the Ledger",
	0x6E00: "open the Cosmos app on the Ledger",
	0x6E01: "open the Cosmos app on the Ledger",
	0x5515: "the Ledger is locked",
}

// StatusError is a non-success status word returned by the Cosmos app.
type StatusError struct {
	Status uint16
	Detail string
}

func (e *StatusError) Error() string {
	msg := fmt.Sprintf("ledger cosmos app: status 0x%04X", e.Status)
	if t, ok := statusText[e.Status]; ok {
		msg += ": " + t
	}
	if e.Detail != "" {
		msg += ": " + e.Detail
	}
	return msg
}

// Transport exchanges a single APDU with the device and returns the response data and status word.
type Transport interface {
	Exchange(apdu []byte) (data []byte, status uint16, err error)
	Close() error
}

// App implements sdkledger.SECP256K1 on top of the Ledger Cosmos app.
type App struct {
	transport Transport
	hrp       string
}

var _ sdkledger.SECP256K1 = (*App)(nil)

// New returns a Cosmos app client that uses hrp for sign requests and public key lookups.
func New(t Transport, hrp string) *App {
	return &App{transport: t, hrp: hrp}
}

// Close releases the device.
func (a *App) Close() error {
	return a.transport.Close()
}

// Version returns the Cosmos app version as major.minor.patch.
func (a *App) Version() (string, error) {
	data, err := a.exchange(insGetVersion, 0, 0, nil)
	if err != nil {
		return "", err
	}
	if len(data) < 4 {
		return "", fmt.Errorf("ledger cosmos app: short version response (%d bytes)", len(data))
	}
	return fmt.Sprintf("%d.%d.%d", data[1], data[2], data[3]), nil
}

// GetPublicKeySECP256K1 returns the compressed public key without asking for confirmation.
func (a *App) GetPublicKeySECP256K1(path []uint32) ([]byte, error) {
	pub, _, err := a.getAddress(path, a.hrp, false)
	return pub, err
}

// GetAddressPubKeySECP256K1 shows the address on the device and returns the public key and address.
// It fails if the address shown on the device is not the eth_secp256k1 address of the public key,
// so that a key is never stored under a different address than the one the user approved.
func (a *App) GetAddressPubKeySECP256K1(path []uint32, hrp string) ([]byte, string, error) {
	pub, shown, err := a.getAddress(path, hrp, true)
	if err != nil {
		return nil, "", err
	}
	if len(pub) != ethsecp256k1.PubKeySize {
		return nil, "", fmt.Errorf("ledger cosmos app: unexpected public key length %d", len(pub))
	}
	want, err := sdk.Bech32ifyAddressBytes(hrp, (&ethsecp256k1.PubKey{Key: pub}).Address())
	if err != nil {
		return nil, "", err
	}
	if shown != want {
		return nil, "", fmt.Errorf("ledger cosmos app: the device shows %s, expected the eth_secp256k1 address %s", shown, want)
	}
	return pub, shown, nil
}

// SignSECP256K1 signs msg and returns a DER signature. p2 is 0 for amino JSON and 1 for textual.
// The first chunk carries the path and the HRP, as the Cosmos app requires on coin type 60.
func (a *App) SignSECP256K1(path []uint32, msg []byte, p2 byte) ([]byte, error) {
	if err := validateHRP(a.hrp); err != nil {
		return nil, err
	}
	first, err := pathBytes(path)
	if err != nil {
		return nil, err
	}
	first = append(first, byte(len(a.hrp))) //nolint:gosec // validateHRP bounds the length
	first = append(first, a.hrp...)

	chunks := [][]byte{first}
	for start := 0; start < len(msg); start += chunkSize {
		chunks = append(chunks, msg[start:min(start+chunkSize, len(msg))])
	}
	if len(chunks) == 1 {
		return nil, errors.New("ledger cosmos app: empty message")
	}

	var sig []byte
	for i, chunk := range chunks {
		p1 := byte(chunkAdd)
		switch i {
		case 0:
			p1 = chunkInit
		case len(chunks) - 1:
			p1 = chunkLast
		}
		if sig, err = a.exchange(insSign, p1, p2, chunk); err != nil {
			return nil, err
		}
	}
	return sig, nil
}

func (a *App) getAddress(path []uint32, hrp string, confirm bool) ([]byte, string, error) {
	if err := validateHRP(hrp); err != nil {
		return nil, "", err
	}
	p, err := pathBytes(path)
	if err != nil {
		return nil, "", err
	}
	data := append([]byte{byte(len(hrp))}, hrp...) //nolint:gosec // validateHRP bounds the length
	data = append(data, p...)

	p1 := byte(0)
	if confirm {
		p1 = 1
	}
	resp, err := a.exchange(insGetAddr, p1, 0, data)
	if err != nil {
		return nil, "", err
	}
	if len(resp) < pubKeyLen {
		return nil, "", fmt.Errorf("ledger cosmos app: short address response (%d bytes)", len(resp))
	}
	return resp[:pubKeyLen], string(resp[pubKeyLen:]), nil
}

func (a *App) exchange(ins, p1, p2 byte, data []byte) ([]byte, error) {
	if len(data) > 255 {
		return nil, fmt.Errorf("ledger cosmos app: APDU data too long (%d bytes)", len(data))
	}
	apdu := append([]byte{cla, ins, p1, p2, byte(len(data))}, data...) //nolint:gosec // checked above
	resp, status, err := a.transport.Exchange(apdu)
	if err != nil {
		return nil, err
	}
	if status != swOK {
		return nil, &StatusError{Status: status, Detail: printable(resp)}
	}
	return resp, nil
}

// validateHRP checks the bech32 HRP length limits.
func validateHRP(hrp string) error {
	if len(hrp) == 0 || len(hrp) > 83 {
		return fmt.Errorf("ledger cosmos app: invalid HRP %q", hrp)
	}
	return nil
}

// pathBytes encodes a 5-level BIP44 path, hardening the first three levels.
func pathBytes(path []uint32) ([]byte, error) {
	if len(path) != 5 {
		return nil, fmt.Errorf("ledger cosmos app: path must have 5 levels, got %d", len(path))
	}
	out := make([]byte, 4*len(path))
	for i, v := range path {
		if i < hardenCount {
			v |= 0x80000000
		}
		binary.LittleEndian.PutUint32(out[4*i:], v)
	}
	return out, nil
}

// printable returns the app's error text, if the response is printable ASCII.
func printable(b []byte) string {
	for _, c := range b {
		if c < 0x20 || c > 0x7e {
			return ""
		}
	}
	return string(b)
}
