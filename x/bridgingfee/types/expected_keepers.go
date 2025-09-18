package types

import (
	"context"

	"github.com/bcp-innovations/hyperlane-cosmos/util"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type BankKeeper interface {
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
}

type WarpQuery interface {
	Token(context.Context, *warptypes.QueryTokenRequest) (*warptypes.QueryTokenResponse, error)
}

type CoreKeeper interface {
	PostDispatchRouter() *util.Router[util.PostDispatchModule]
}

type TxFeesKeeper interface {
	CalcCoinInBaseDenom(ctx sdk.Context, inputFee sdk.Coin) (sdk.Coin, error)
}
