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

// consumeMinimumGas consumes the minimum gas
// if the consumed gas during tx is less than the minimum gas.
func consumeMinimumGas(ctx sdk.Context, minimumGas uint64, actionName string) {
	if minimumGas > 0 {
		if consumedGas := ctx.GasMeter().GasConsumed(); consumedGas < minimumGas {
			// we only consume the gas that is needed to reach the minimum gas
			gasToConsume := minimumGas - consumedGas

			ctx.GasMeter().ConsumeGas(gasToConsume, actionName)
		}
	}
}
