package utils

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/app/params"
)

func TestCoins(amount int64) sdk.Coins {
	return sdk.Coins{
		{
			Denom:  params.BaseDenom,
			Amount: sdk.NewInt(amount),
		},
	}
}

func TestCoin(amount int64) sdk.Coin {
	return sdk.Coin{
		Denom:  params.BaseDenom,
		Amount: sdk.NewInt(amount),
	}
}

func TestCoinP(amount int64) *sdk.Coin {
	coin := TestCoin(amount)
	return &coin
}

func TestCoin2P(coin sdk.Coin) *sdk.Coin {
	return &coin
}
