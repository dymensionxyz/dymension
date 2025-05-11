package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
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
// The original consumed gas should be captured from gas meter before invoke message execution.
// This function will panic if the gas meter consumed gas is less than the original consumed gas.
func consumeMinimumGas(ctx sdk.Context, minimumGas, originalConsumedGas uint64, actionName string) {
	if minimumGas > 0 {
		laterConsumedGas := ctx.GasMeter().GasConsumed()
		if laterConsumedGas < originalConsumedGas {
			// unexpect gas consumption
			panic(fmt.Sprintf(
				"later gas is less than original gas: %d < %d",
				laterConsumedGas, originalConsumedGas,
			))
		}
		if consumedGas := laterConsumedGas - originalConsumedGas; consumedGas < minimumGas {
			// we only consume the gas that is needed to reach the target minimum gas
			gasToConsume := minimumGas - consumedGas

			ctx.GasMeter().ConsumeGas(gasToConsume, actionName)
		}
	}
}

// MigrateChainIds implements the Msg/MigrateChainIds RPC method
func (k msgServer) MigrateChainIds(goCtx context.Context, msg *dymnstypes.MsgMigrateChainIds) (*dymnstypes.MsgMigrateChainIdsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	if msg.Authority != k.authority {
		return nil, errorsmod.Wrapf(gerrc.ErrUnauthenticated, "invalid authority; expected %s, got %s", k.authority, msg.Authority)
	}

	if err := k.Keeper.MigrateChainIds(ctx, msg.Replacement); err != nil {
		return nil, err
	}

	return &dymnstypes.MsgMigrateChainIdsResponse{}, nil
}

// UpdateAliases implements the Msg/UpdateAliases RPC method
func (k msgServer) UpdateAliases(goCtx context.Context, msg *dymnstypes.MsgUpdateAliases) (*dymnstypes.MsgUpdateAliasesResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	if msg.Authority != k.authority {
		return nil, errorsmod.Wrapf(gerrc.ErrUnauthenticated, "invalid authority; expected %s, got %s", k.authority, msg.Authority)
	}

	if err := k.Keeper.UpdateAliases(ctx, msg.Add, msg.Remove); err != nil {
		return nil, err
	}

	return &dymnstypes.MsgUpdateAliasesResponse{}, nil
}
