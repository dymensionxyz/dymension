package types

import (
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
