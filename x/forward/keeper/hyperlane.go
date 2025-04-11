package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	hyperutil "github.com/bcp-innovations/hyperlane-cosmos/util"
	warpkeeper "github.com/bcp-innovations/hyperlane-cosmos/x/warp/keeper"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/forward/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// for inbound warp route transfers. At this point, the tokens are in the hyperlane warp module still
func (k Keeper) OnHyperlaneMessage(goCtx context.Context, args warpkeeper.OnHyperlaneMessageArgs) {
	// TODO: should allow another level of indirection (e.g. Memo is json containing what we want in bytes?)
	// it would be more flexible and allow memo forwarding

	ctx := sdk.UnwrapSDKContext(goCtx)

	k.refundOnError(ctx, func() error {

		d, nextMemo, err := types.UnpackAppMemoFromHyperlaneMemo(args.Memo)
		if err != nil {
			return errorsmod.Wrap(err, "unpack memo from hyperlane")
		}

		return k.forwardToIBC(ctx, d.Transfer, args.Coin(), nextMemo)
	}, nil, warptypes.ModuleName, args.Account, args.Coin())
}

func (k Keeper) forwardToHyperlane(ctx sdk.Context, fundsSrc sdk.AccAddress, budget sdk.Coin, d types.HookEIBCtoHL) error {

	token, err := k.getHypToken(ctx, hyperutil.HexAddress(d.HyperlaneTransfer.TokenId))
	if err != nil {
		return errorsmod.Wrap(err, "get hyp token")
	}

	if token.OriginDenom != budget.Denom {
		return gerrc.ErrInvalidArgument.Wrapf("token denom does not match allowed denom: %s != %s", token.OriginDenom, budget.Denom)
	}
	if d.HyperlaneTransfer.MaxFee.Denom != budget.Denom {
		return gerrc.ErrInvalidArgument.Wrapf("max fee denom does not match allowed denom: %s != %s", d.HyperlaneTransfer.MaxFee.Denom, budget.Denom)
	}
	maxCost := d.HyperlaneTransfer.MaxFee.Amount.Add(d.HyperlaneTransfer.Amount)
	if maxCost.GT(budget.Amount) {
		return gerrc.ErrInvalidArgument.Wrapf("max cost (fee + amount)exceeds max budget %s > %s", maxCost, budget.Amount)
	}

	// Need to use DymRemoteTransfer because we only use WR's which support memo in the direction HL -> Hub, and
	// we need to send back with the same WR that it came in on.
	m := &warptypes.MsgDymRemoteTransfer{
		Inner: &warptypes.MsgRemoteTransfer{

			Sender:            fundsSrc.String(),
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
		},
	}

	_, err = k.warpS.DymRemoteTransfer(ctx, m) // TODO: responsse?
	err = errorsmod.Wrap(err, "dym remote transfer")
	return err

}

func (k Keeper) getHypToken(ctx context.Context, tokenId hyperutil.HexAddress) (*warptypes.WrappedHypToken, error) {
	res, err := k.warpQ.Token(ctx, &warptypes.QueryTokenRequest{Id: tokenId.String()})
	if err != nil {
		return nil, err
	}

	return res.Token, nil
}
