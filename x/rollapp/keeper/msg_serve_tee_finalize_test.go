package keeper

import (
	"strings"
	"testing"

	_ "embed"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

//go:embed testdata/confidential_space_root.pem
var testPEM []byte

// openssl x509 -fingerprint -in confidential_space_root.pem -noout
// https://codelabs.developers.google.com/confidential-space-pki?hl=en#2
func TestValidatePEMCert(t *testing.T) {
	expect := "B9:51:20:74:2C:24:E3:AA:34:04:2E:1C:3B:A3:AA:D2:8B:21:23:21"
	expect = strings.ReplaceAll(expect, ":", "")
	err := validatePEMCert(sdk.Context{}, testPEM, expect)
	if err != nil {
		t.Fatalf("validatePEMCert: %v", err)
	}
}
