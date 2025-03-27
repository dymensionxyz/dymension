package types

import (
	context "context"

	"cosmossdk.io/math"
	"github.com/bcp-innovations/hyperlane-cosmos/util"
	"github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type WarpRouteKeeper interface {
	RemoteTransferCollateral(ctx sdk.Context,
		token types.HypToken,
		cosmosSender string,
		destinationDomain uint32,
		externalRecipient util.HexAddress,
		amount math.Int,
		customHookId *util.HexAddress,
		gasLimit math.Int,
		maxFee sdk.Coin,
		customHookMetadata []byte) (messageId util.HexAddress, err error)
}

type WarpQuery interface {
	Token(ctx context.Context, request *types.QueryTokenRequest) (*types.QueryTokenResponse, error)
}

type BankKeeper interface {
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx context.Context, senderModule, recipientModule string, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

type AccountKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
}
