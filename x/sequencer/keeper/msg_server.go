package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/sequencer/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

// Bond defines a method for adding coins to sequencer's bond
func (k msgServer) Bond(context.Context, *types.MsgBond) (*types.MsgBondResponse, error) {
	panic("implement me")
}

// Unbond defines a method for removing coins from sequencer's bond
func (k msgServer) Unbond(goCtx context.Context, msg *types.MsgUnbond) (*types.MsgUnbondResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	completionTime, err := k.StartUnbondingSequencer(ctx, msg.Creator)
	if err != nil {
		return nil, err
	}

	//TODO: emit events
	// ctx.EventManager().EmitEvents(sdk.Events{
	// 	sdk.NewEvent(
	// 		types.EventTypeUnbond,
	// 		sdk.NewAttribute(types.AttributeKeyValidator, msg.ValidatorAddress),
	// 		sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
	// 		sdk.NewAttribute(types.AttributeKeyCompletionTime, completionTime.Format(time.RFC3339)),
	// 	),
	// 	sdk.NewEvent(
	// 		sdk.EventTypeMessage,
	// 		sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
	// 		sdk.NewAttribute(sdk.AttributeKeySender, msg.DelegatorAddress),
	// 	),
	// })

	return &types.MsgUnbondResponse{
		CompletionTime: completionTime,
	}, nil
}
