package types

import sdk "github.com/cosmos/cosmos-sdk/types"

type StubGammHooks struct{}

func (StubGammHooks) AfterJoinPool(sdk.Context, sdk.AccAddress, uint64, sdk.Coins, sdk.Int) {}
func (StubGammHooks) AfterExitPool(sdk.Context, sdk.AccAddress, uint64, sdk.Int, sdk.Coins) {}
func (StubGammHooks) AfterSwap(sdk.Context, sdk.AccAddress, uint64, sdk.Coins, sdk.Coins)   {}
