package uibc

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: this test is just to keep the sdk imported, until we need it, because it's annoying to add to the go.mod

func TestKeepSDK(t *testing.T) {
	t.Log(sdk.Context{})
}
