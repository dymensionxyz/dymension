package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
	testkeeper "github.com/dymensionxyz/dymension/v3/testutil/keeper"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	dymnsutils "github.com/dymensionxyz/dymension/v3/x/dymns/utils"
	"github.com/stretchr/testify/require"
)

//goland:noinspection SpellCheckingInspection
func TestKeeper_RefundBid(t *testing.T) {
	dymNsModuleAccAddr := authtypes.NewModuleAddress(dymnstypes.ModuleName)

	tests := []struct {
		name                     string
		refundToAccount          string
		refundAmount             sdk.Coin
		fundModuleAccountBalance sdk.Coin
		genesis                  bool
		wantErr                  bool
		wantErrContains          string
	}{
		{
			name:            "refund bid normally but account has no balance",
			refundToAccount: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			refundAmount:    dymnsutils.TestCoin(100),
			genesis:         false,
			wantErr:         true,
			wantErrContains: "insufficient funds",
		},
		{
			name:                     "refund bid normally but account does not have enough balance",
			refundToAccount:          "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			refundAmount:             dymnsutils.TestCoin(100),
			fundModuleAccountBalance: dymnsutils.TestCoin(50),
			genesis:                  false,
			wantErr:                  true,
			wantErrContains:          "insufficient funds",
		},
		{
			name:                     "refund bid normally",
			refundToAccount:          "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			refundAmount:             dymnsutils.TestCoin(100),
			fundModuleAccountBalance: dymnsutils.TestCoin(150),
			genesis:                  false,
		},
		{
			name:            "refund bid genesis",
			refundToAccount: "dym1fl48vsnmsdzcv85q5d2q4z5ajdha8yu38x9fue",
			refundAmount:    dymnsutils.TestCoin(100),
			genesis:         true,
			wantErr:         false,
		},
		{
			name:                     "bad bidder",
			refundToAccount:          "dym1fl48vsnmsdzcv85q5d",
			refundAmount:             dymnsutils.TestCoin(100),
			fundModuleAccountBalance: dymnsutils.TestCoin(100),
			wantErr:                  true,
			wantErrContains:          "SO bidder is not a valid bech32 account address",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dk, bk, _, ctx := testkeeper.DymNSKeeper(t)

			if !tt.fundModuleAccountBalance.IsNil() {
				if !tt.fundModuleAccountBalance.IsZero() {
					err := bk.MintCoins(ctx, dymnstypes.ModuleName, sdk.Coins{tt.fundModuleAccountBalance})
					require.NoError(t, err)
				}
			}

			soBid := dymnstypes.SellOrderBid{
				Bidder: tt.refundToAccount,
				Price:  tt.refundAmount,
			}

			var err error
			if tt.genesis {
				err = dk.GenesisRefundBid(ctx, soBid)
			} else {
				err = dk.RefundBid(ctx, soBid)
			}

			if tt.wantErr {
				require.NotEmpty(t, tt.wantErrContains, "mis-configured test case")
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrContains)
				return
			}

			require.NoError(t, err)

			laterBidderBalance := bk.GetBalance(ctx, sdk.MustAccAddressFromBech32(tt.refundToAccount), params.BaseDenom)
			require.Equal(t, tt.refundAmount.Amount.BigInt(), laterBidderBalance.Amount.BigInt())

			laterDymNsModuleBalance := bk.GetBalance(ctx, dymNsModuleAccAddr, params.BaseDenom)
			if tt.genesis {
				require.True(t, laterDymNsModuleBalance.IsZero())
			} else {
				require.Equal(t, tt.fundModuleAccountBalance.Sub(tt.refundAmount).Amount.BigInt(), laterDymNsModuleBalance.Amount.BigInt())
			}

			// event should be fired
			events := ctx.EventManager().Events()
			require.NotEmpty(t, events)

			var found bool
			for _, event := range events {
				if event.Type == dymnstypes.EventTypeDymNameRefundBid {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("event %s not found", dymnstypes.EventTypeDymNameRefundBid)
			}
		})
	}
}
