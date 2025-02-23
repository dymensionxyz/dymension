package types

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type StubGammHooks struct{}

func (StubGammHooks) AfterJoinPool(sdk.Context, sdk.AccAddress, uint64, sdk.Coins, math.Int) {}
func (StubGammHooks) AfterExitPool(sdk.Context, sdk.AccAddress, uint64, math.Int, sdk.Coins) {}
func (StubGammHooks) AfterSwap(sdk.Context, sdk.AccAddress, uint64, sdk.Coins, sdk.Coins)    {}
