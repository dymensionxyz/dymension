package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	warpkeeper "github.com/bcp-innovations/hyperlane-cosmos/x/warp/keeper"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	eibctypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
	"github.com/dymensionxyz/dymension/v3/x/forward/types"
)

// for inbound warp route transfers. At this point, the tokens are in the hyperlane warp module still
func (k Keeper) OnHyperlane(goCtx context.Context, args warpkeeper.DymHookArgs) error {
	// TODO: should allow another level of indirection (e.g. Memo is json containing what we want in bytes?)
	// it would be more flexible and allow memo forwarding
	var d types.HookHLtoIBC
	err := proto.Unmarshal(args.Memo, &d)
	if err != nil {
		return errorsmod.Wrap(err, "unmarshal forward hook")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	err = k.escrowFromModule(ctx, warptypes.ModuleName, args.Coins)
	if err != nil {
		// should never happen
		err = errorsmod.Wrap(err, "escrow from module")
		k.Logger(ctx).Error("onHyperlane", "error", err)
		return err
	}

	k.forwardToIBC(ctx, d.Transfer, *d.Recovery)
	return nil
}

// for transfers coming from eibc which are being forwarded (to HL)
func (k Keeper) forwardToHyperlane(ctx sdk.Context, order *eibctypes.DemandOrder, d types.HookEIBCtoHL) {

	// TODO: anything to change?
	m := &warptypes.MsgRemoteTransfer{
		Sender:             d.HyperlaneTransfer.Sender,
		TokenId:            d.HyperlaneTransfer.TokenId,
		DestinationDomain:  d.HyperlaneTransfer.DestinationDomain,
		Recipient:          d.HyperlaneTransfer.Recipient,
		Amount:             d.HyperlaneTransfer.Amount,
		CustomHookId:       d.HyperlaneTransfer.CustomHookId,
		GasLimit:           d.HyperlaneTransfer.GasLimit,
		MaxFee:             d.HyperlaneTransfer.MaxFee,
		CustomHookMetadata: d.HyperlaneTransfer.CustomHookMetadata,
	}

	k.refundOnError(ctx, func() error {
		_, err := k.warpServer.RemoteTransfer(ctx, m) // TODO: responsse?
		return err
	}, *d.Recovery, order.Price)

}
