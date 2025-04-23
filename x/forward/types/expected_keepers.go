package types

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/hyperlane-cosmos/x/warp/types"
)

type WarpQuery interface {
	Token(ctx context.Context, request *types.QueryTokenRequest) (*types.QueryTokenResponse, error)
}

type BankKeeper interface {
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoins(ctx context.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
}

type AccountKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
}
