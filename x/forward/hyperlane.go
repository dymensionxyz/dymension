package forward

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/forward/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	hyperutil "github.com/dymensionxyz/hyperlane-cosmos/util"
	warpkeeper "github.com/dymensionxyz/hyperlane-cosmos/x/warp/keeper"
	warptypes "github.com/dymensionxyz/hyperlane-cosmos/x/warp/types"
)

// this is called by hyperlane on an inbound transfer
// at time of calling, funds have already been credited to the original hyperlane transfer recipient
func (k Forward) OnHyperlaneMessage(goCtx context.Context, args warpkeeper.OnHyperlaneMessageArgs) error {

	ctx := sdk.UnwrapSDKContext(goCtx)

	if len(args.Memo) == 0 {
		// Equivalent to the vanilla token standard. Might be used to provide EIBC or LP funds.
		return nil
	}

	// if it fails, the original hyperlane transfer recipient got the funds anyway so no need to do anything special (relying on frontend here)
	k.executeWithErrEvent(ctx, func() error {
		d, err := types.UnpackForwardToIBC(args.Memo)
		if err != nil {
			return errorsmod.Wrap(err, "unpack memo from hyperlane")
		}

		// funds src is the hyperlane transfer recipient, which should have same priv key as rollapp recipient
		// so in case of async failure, the funds will get refunded back there.
		return k.forwardToIBC(ctx, d.Transfer, args.Account, args.Coin())
	})

	return nil
}

func (k Forward) forwardToHyperlane(ctx sdk.Context, fundsSrc sdk.AccAddress, budget sdk.Coin, d types.HookForwardToHL) error {

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

func (k Forward) getHypToken(ctx context.Context, tokenId hyperutil.HexAddress) (*warptypes.WrappedHypToken, error) {
	res, err := k.warpQ.Token(ctx, &warptypes.QueryTokenRequest{Id: tokenId.String()})
	if err != nil {
		return nil, err
	}

	return res.Token, nil
}
