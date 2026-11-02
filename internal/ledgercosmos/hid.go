package ledgercosmos

import (
	"encoding/binary"
	"errors"
	"fmt"
	"sync"
	"time"

	sdkledger "github.com/cosmos/cosmos-sdk/crypto/ledger"
	"github.com/zondax/hid"
)

const (
	ledgerVendorID = 0x2c97
	ledgerUsage    = 0xffa0
	packetSize     = 64
	channel        = 0x0101
	tagAPDU        = 0x05
)

// HID is a USB HID transport for Ledger devices.
//
// Each 64-byte report is written with its own write call. ledger-go writes all
// reports of an APDU in a single call; on macOS that write fails on the last chunk of a
// sign request ("hidapi: unknown failure") and leaves the device unresponsive until it is
// reconnected.
type HID struct {
	device    *hid.Device
	reads     chan []byte
	done      chan struct{}
	closeOnce sync.Once
}

var _ Transport = (*HID)(nil)

// OpenHID opens the first connected Ledger.
func OpenHID() (*HID, error) {
	if !hid.Supported() {
		return nil, errors.New("ledger: USB HID is not supported by this build (cgo is required)")
	}

	var info *hid.DeviceInfo
	for _, d := range hid.Enumerate(ledgerVendorID, 0) {
		// macOS and Windows report the usage page; Linux reports the interface number.
		if d.UsagePage == ledgerUsage || (d.UsagePage == 0 && d.Interface == 0) {
			info = &d
			break
		}
	}
	if info == nil {
		return nil, errors.New("ledger: no device found (is it connected, unlocked and not in use by another program?)")
	}

	device, err := info.Open()
	if err != nil {
		return nil, fmt.Errorf("ledger: open: %w", err)
	}

	t := &HID{device: device, reads: make(chan []byte, 16), done: make(chan struct{})}
	go t.readLoop()
	return t, nil
}

func (t *HID) readLoop() {
	for {
		buf := make([]byte, packetSize)
		n, err := t.device.Read(buf)
		if err != nil {
			if errors.Is(err, hid.ErrDeviceClosed) {
				close(t.reads)
				return
			}
			select {
			case <-t.done:
				close(t.reads)
				return
			case <-time.After(10 * time.Millisecond):
				continue
			}
		}
		select {
		case t.reads <- buf[:n]:
		case <-t.done:
			close(t.reads)
			return
		}
	}
}

// Exchange sends one APDU and returns the response data and status word.
func (t *HID) Exchange(apdu []byte) ([]byte, uint16, error) {
	for _, p := range frame(apdu) {
		if _, err := t.device.Write(p); err != nil {
			return nil, 0, fmt.Errorf("ledger: write: %w", err)
		}
	}

	resp, err := unframe(t.reads)
	if err != nil {
		return nil, 0, err
	}
	if len(resp) < 2 {
		return nil, 0, fmt.Errorf("ledger: short response (%d bytes)", len(resp))
	}
	n := len(resp) - 2
	return resp[:n], binary.BigEndian.Uint16(resp[n:]), nil
}

// Close closes the device.
func (t *HID) Close() error {
	var err error
	t.closeOnce.Do(func() {
		close(t.done)
		err = t.device.Close()
	})
	return err
}

// frame splits an APDU into Ledger HID reports: channel, tag, sequence index and,
// in the first report, the APDU length.
func frame(apdu []byte) [][]byte {
	data := make([]byte, 2+len(apdu))
	binary.BigEndian.PutUint16(data, uint16(len(apdu))) //nolint:gosec // APDUs are at most 260 bytes
	copy(data[2:], apdu)

	var reports [][]byte
	for seq := 0; len(data) > 0; seq++ {
		r := make([]byte, packetSize)
		binary.BigEndian.PutUint16(r[0:], channel)
		r[2] = tagAPDU
		binary.BigEndian.PutUint16(r[3:], uint16(seq)) //nolint:gosec // a few reports per APDU
		data = data[copy(r[5:], data):]
		reports = append(reports, r)
	}
	return reports
}

// unframe reads reports until a whole response has arrived.
func unframe(reads <-chan []byte) ([]byte, error) {
	var resp []byte
	total := -1
	for seq := 0; total < 0 || len(resp) < total; seq++ {
		r, ok := <-reads
		if !ok {
			return nil, errors.New("ledger: device closed")
		}
		if len(r) < 5 || binary.BigEndian.Uint16(r[0:]) != channel || r[2] != tagAPDU ||
			int(binary.BigEndian.Uint16(r[3:])) != seq {
			return nil, fmt.Errorf("ledger: unexpected report %x", r)
		}
		data := r[5:]
		if seq == 0 {
			if len(data) < 2 {
				return nil, errors.New("ledger: short first report")
			}
			total = int(binary.BigEndian.Uint16(data))
			data = data[2:]
		}
		resp = append(resp, data...)
	}
	return resp[:total], nil
}

// Discover returns a sdkledger derivation function that connects to the first Ledger over USB.
// hrp is resolved on every call so that it follows the SDK bech32 configuration.
func Discover(hrp func() string) func() (sdkledger.SECP256K1, error) {
	return func() (sdkledger.SECP256K1, error) {
		t, err := OpenHID()
		if err != nil {
			return nil, err
		}
		app := New(t, hrp())
		if _, err := app.Version(); err != nil {
			_ = t.Close()
			return nil, err
		}
		return app, nil
	}
}
