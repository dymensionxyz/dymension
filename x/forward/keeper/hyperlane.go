package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	hyperutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	warpkeeper "github.com/bcp-innovations/hyperlane-cosmos/x/warp/keeper"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"
	eibctypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
	"github.com/dymensionxyz/dymension/v3/x/forward/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
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

	k.forwardToIBC(ctx, d.Transfer, *d.Recovery, args.Coins[0])
	return nil
}

func (k Keeper) forwardToHyperlane(ctx sdk.Context, order *eibctypes.DemandOrder, d types.HookEIBCtoHL) {
	k.refundOnError(ctx, func() error {
		return k.forwardToHyperlaneInner(ctx, order, d)
	}, *d.Recovery, order.Price)
}

func (k Keeper) forwardToHyperlaneInner(ctx sdk.Context, order *eibctypes.DemandOrder, d types.HookEIBCtoHL) error {

	maxBudget := order.PriceAmount()
	allowedDenom := order.Denom()

	token, err := k.getHypToken(ctx, hyperutil.HexAddress(d.HyperlaneTransfer.TokenId))
	if err != nil {
		return errorsmod.Wrap(err, "get hyp token")
	}

	if token.OriginDenom != allowedDenom {
		return gerrc.ErrInvalidArgument.Wrapf("token denom does not match allowed denom: %s != %s", token.OriginDenom, allowedDenom)
	}
	if d.HyperlaneTransfer.MaxFee.Denom != allowedDenom {
		return gerrc.ErrInvalidArgument.Wrapf("max fee denom does not match allowed denom: %s != %s", d.HyperlaneTransfer.MaxFee.Denom, allowedDenom)
	}
	maxCost := d.HyperlaneTransfer.MaxFee.Amount.Add(d.HyperlaneTransfer.Amount)
	if maxCost.GT(maxBudget) {
		return gerrc.ErrInvalidArgument.Wrapf("max cost (fee + amount)exceeds max budget %s > %s", maxCost, maxBudget)
	}

	m := &warptypes.MsgRemoteTransfer{
		Sender:            k.getModuleAddr(ctx).String(),
		TokenId:           d.HyperlaneTransfer.TokenId,
		DestinationDomain: d.HyperlaneTransfer.DestinationDomain,
		Recipient:         d.HyperlaneTransfer.Recipient,
		Amount:            d.HyperlaneTransfer.Amount,

		GasLimit: d.HyperlaneTransfer.GasLimit,
		MaxFee:   d.HyperlaneTransfer.MaxFee,

		// these are used mainly to override the default relayer (pay a different relayer)
		// there is no security risk from allowing these to be anything
		CustomHookMetadata: d.HyperlaneTransfer.CustomHookMetadata,
		CustomHookId:       d.HyperlaneTransfer.CustomHookId,
	}

	_, err = k.warpS.RemoteTransfer(ctx, m) // TODO: responsse?
	return err

}

func (k Keeper) getHypToken(ctx context.Context, tokenId hyperutil.HexAddress) (*warptypes.WrappedHypToken, error) {
	res, err := k.warpQ.Token(ctx, &warptypes.QueryTokenRequest{Id: tokenId.String()})
	if err != nil {
		return nil, err
	}

	return res.Token, nil
}
