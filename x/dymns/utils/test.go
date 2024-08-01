package utils

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
)

// TestCoins is used for testing purposes.
func TestCoins(amount int64) sdk.Coins {
	return sdk.Coins{
		{
			Denom:  params.BaseDenom,
			Amount: sdk.NewInt(amount),
		},
	}
}

// TestCoin is used for testing purposes.
func TestCoin(amount int64) sdk.Coin {
	return sdk.Coin{
		Denom:  params.BaseDenom,
		Amount: sdk.NewInt(amount),
	}
}

// TestCoinP is used for testing purposes.
func TestCoinP(amount int64) *sdk.Coin {
	coin := TestCoin(amount)
	return &coin
}

// TestCoin2P is used for testing purposes.
func TestCoin2P(coin sdk.Coin) *sdk.Coin {
	return &coin
}
