package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

var _ dymnstypes.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) dymnstypes.MsgServer {
	return &msgServer{Keeper: keeper}
}

func consumeMinimumGas(ctx sdk.Context, minimumGas uint64, actionName string) {
	if minimumGas > 0 {
		if consumedGas := ctx.GasMeter().GasConsumed(); consumedGas < minimumGas {
			ctx.GasMeter().ConsumeGas(minimumGas-consumedGas, actionName)
		}
	}
}
