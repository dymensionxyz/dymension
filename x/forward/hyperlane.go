package forward

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

// this is called by hyperlane on an inbound transfer
// at time of calling, funds have already been credited to the original hyperlane transfer recipient
func (k Forward) OnHyperlaneMessage(goCtx context.Context, args warpkeeper.OnHyperlaneMessageArgs) error {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// if it fails, the original hyperlane transfer recipient got the funds anyway so no need to do anything special (relying on frontend here)
	k.executeWithErrEvent(ctx, func() (bool, error) {
		hlMetadata, err := types.UnpackHLMetadata(args.Metadata)
		if err != nil {
			return false, errorsmod.Wrap(err, "unpack hl metadata")
		}
		if hlMetadata == nil {
			// Equivalent to the vanilla token standard.
			return false, nil
		}

		// Check for HL-to-HL forwarding first
		if len(hlMetadata.HookForwardToHl) > 0 {
			d, err := types.UnpackForwardToHL(hlMetadata.HookForwardToHl)
			if err != nil {
				return true, errorsmod.Wrap(err, "unpack hl to hl forward from hyperlane")
			}

			// funds src is the hyperlane transfer recipient
			return true, k.forwardToHyperlane(ctx, args.Account, args.Coin(), *d)
		}

		// Check for HL-to-IBC forwarding
		if len(hlMetadata.HookForwardToIbc) > 0 {
			d, err := types.UnpackForwardToIBC(hlMetadata.HookForwardToIbc)
			if err != nil {
				return true, errorsmod.Wrap(err, "unpack memo from hyperlane")
			}

			// funds src is the hyperlane transfer recipient, which should have same priv key as rollapp recipient
			// so in case of async failure, the funds will get refunded back there.
			return true, k.forwardToIBC(ctx, d.Transfer, args.Account, args.Coin())
		}

		// No forwarding configured
		return false, nil
	})

	return nil
}

func (k Forward) forwardToHyperlane(ctx sdk.Context, fundsSrc sdk.AccAddress, budget sdk.Coin, d types.HookForwardToHL) error {
	token, err := k.getHypToken(ctx, d.HyperlaneTransfer.TokenId)
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

	m := &warptypes.MsgRemoteTransfer{
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
	}

	_, err = k.warpS.RemoteTransfer(ctx, m) // TODO: responsse?
	return errorsmod.Wrap(err, "dym remote transfer")
}

func (k Forward) getHypToken(ctx context.Context, tokenId hyperutil.HexAddress) (*warptypes.WrappedHypToken, error) {
	res, err := k.warpQ.Token(ctx, &warptypes.QueryTokenRequest{Id: tokenId.String()})
	if err != nil {
		return nil, err
	}

	return res.Token, nil
}
