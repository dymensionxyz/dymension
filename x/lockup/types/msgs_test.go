package types_test

import (
	"testing"
	"time"

	"cosmossdk.io/math"
	"github.com/cometbft/cometbft/crypto/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/dymensionxyz/dymension/v3/app/apptesting"
	"github.com/dymensionxyz/dymension/v3/x/lockup/types"
)

func TestMsgBeginUnlocking(t *testing.T) {
	addr1 := apptesting.CreateRandomAccounts(1)[0].String()
	invalidAddr := sdk.AccAddress("invalid").String()

	tests := []struct {
		name       string
		msg        types.MsgBeginUnlocking
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: types.MsgBeginUnlocking{
				Owner: addr1,
				ID:    1,
				Coins: sdk.NewCoins(sdk.NewCoin("test", math.NewInt(100))),
			},
			expectPass: true,
		},
		{
			name: "invalid owner",
			msg: types.MsgBeginUnlocking{
				Owner: invalidAddr,
				ID:    1,
				Coins: sdk.NewCoins(sdk.NewCoin("test", math.NewInt(100))),
			},
		},
		{
			name: "invalid lockup ID",
			msg: types.MsgBeginUnlocking{
				Owner: addr1,
				ID:    0,
				Coins: sdk.NewCoins(sdk.NewCoin("test", math.NewInt(100))),
			},
		},
		{
			name: "invalid coins length",
			msg: types.MsgBeginUnlocking{
				Owner: addr1,
				ID:    1,
				Coins: sdk.NewCoins(sdk.NewCoin("test1", math.NewInt(100000)), sdk.NewCoin("test2", math.NewInt(100000))),
			},
		},
		{
			name: "zero coins (same as nil)",
			msg: types.MsgBeginUnlocking{
				Owner: addr1,
				ID:    1,
				Coins: sdk.NewCoins(sdk.NewCoin("test1", math.NewInt(0))),
			},
			expectPass: true,
		},
		{
			name: "nil coins (unlock by ID)",
			msg: types.MsgBeginUnlocking{
				Owner: addr1,
				ID:    1,
				Coins: sdk.NewCoins(),
			},
			expectPass: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.expectPass {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
			} else {
				require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
			}
		})
	}
}

func TestMsgExtendLockup(t *testing.T) {
	addr1 := apptesting.CreateRandomAccounts(1)[0].String()
	invalidAddr := sdk.AccAddress("invalid").String()

	tests := []struct {
		name       string
		msg        types.MsgExtendLockup
		expectPass bool
	}{
		{
			name: "proper msg",
			msg: types.MsgExtendLockup{
				Owner:    addr1,
				ID:       1,
				Duration: time.Hour,
			},
			expectPass: true,
		},
		{
			name: "invalid owner",
			msg: types.MsgExtendLockup{
				Owner:    invalidAddr,
				ID:       1,
				Duration: time.Hour,
			},
		},
		{
			name: "invalid lockup ID",
			msg: types.MsgExtendLockup{
				Owner:    addr1,
				ID:       0,
				Duration: time.Hour,
			},
		},
		{
			name: "invalid duration",
			msg: types.MsgExtendLockup{
				Owner:    addr1,
				ID:       1,
				Duration: -1,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.expectPass {
				require.NoError(t, test.msg.ValidateBasic(), "test: %v", test.name)
			} else {
				require.Error(t, test.msg.ValidateBasic(), "test: %v", test.name)
			}
		})
	}
}

// // Test authz serialize and de-serializes for lockup msg.
func TestAuthzMsg(t *testing.T) {
	app := apptesting.Setup(t)

	pk1 := ed25519.GenPrivKey().PubKey()
	addr1 := sdk.AccAddress(pk1.Address()).String()
	coin := sdk.NewCoin("denom", math.NewInt(1))

	testCases := []struct {
		name string
		msg  sdk.Msg
	}{
		{
			name: "MsgLockTokens",
			msg: &types.MsgLockTokens{
				Owner:    addr1,
				Duration: time.Hour,
				Coins:    sdk.NewCoins(coin),
			},
		},
		{
			name: "MsgBeginUnlocking",
			msg: &types.MsgBeginUnlocking{
				Owner: addr1,
				ID:    1,
				Coins: sdk.NewCoins(coin),
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			apptesting.TestMessageAuthzSerialization(t, app.AppCodec(), tc.msg)
		})
	}
}
