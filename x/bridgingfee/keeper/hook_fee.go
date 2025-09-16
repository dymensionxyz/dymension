package keeper

import (
	"github.com/bcp-innovations/hyperlane-cosmos/util"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) Exists(ctx context.Context, hookId util.HexAddress) (bool, error) {
	//TODO implement me
	panic("implement me")
}

func (k Keeper) PostDispatch(ctx context.Context, mailboxId, hookId util.HexAddress, metadata util.StandardHookMetadata, message util.HyperlaneMessage, maxFee sdk.Coins) (sdk.Coins, error) {
	//TODO implement me
	panic("implement me")
}

func (k Keeper) QuoteDispatch(ctx context.Context, mailboxId, hookId util.HexAddress, metadata util.StandardHookMetadata, message util.HyperlaneMessage) (sdk.Coins, error) {
	//TODO implement me
	panic("implement me")
}

func (k Keeper) HookType() uint8 {
	//TODO implement me
	panic("implement me")
}
