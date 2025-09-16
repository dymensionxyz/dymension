package keeper

import (
	"context"
	"fmt"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	hyputil "github.com/bcp-innovations/hyperlane-cosmos/util"
	postdispatchtypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/02_post_dispatch/types"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	"github.com/dymensionxyz/dymension/v3/x/bridgingfee/types"
)

// FeeHookHandler implements the fee collection post-dispatch hook
type FeeHookHandler struct {
	k Keeper
}

var _ hyputil.PostDispatchModule = FeeHookHandler{}

func (f FeeHookHandler) Exists(ctx context.Context, hookId hyputil.HexAddress) (bool, error) {
	has, err := f.k.feeHooks.Has(ctx, hookId.GetInternalId())
	if err != nil {
		return false, err
	}
	return has, nil
}

func (f FeeHookHandler) HookType() uint8 {
	return postdispatchtypes.POST_DISPATCH_HOOK_TYPE_PROTOCOL_FEE
}

// PostDispatch collects fees from the sender for bridging tokens
func (f FeeHookHandler) PostDispatch(ctx context.Context, _, hookId hyputil.HexAddress, metadata hyputil.StandardHookMetadata, message hyputil.HyperlaneMessage, _ sdk.Coins) (sdk.Coins, error) {
	fee, err := f.quoteFee(ctx, hookId, message.Sender, message.Body)
	if err != nil {
		return nil, err
	}

	if fee.IsZero() {
		return sdk.NewCoins(), nil
	}

	// TODO: think what to do with maxFee denom and fee denom
	// maxFee in MsgRemoteTransfer is Coin (singular) and most likely is in DYM
	// fee is in token.OriginDenom, eg in KAS

	// For now, we don't check maxFee since denominations may differ
	// In production, proper validation should be implemented

	// Collect fee from sender to module account
	sdkCtx := sdk.UnwrapSDKContext(ctx)
	feeCoins := sdk.NewCoins(fee)
	if err := f.k.bankKeeper.SendCoinsFromAccountToModule(sdkCtx, metadata.Address, types.ModuleName, feeCoins); err != nil {
		return nil, errorsmod.Wrap(err, "collect fees")
	}

	err = uevent.EmitTypedEvent(sdk.UnwrapSDKContext(ctx), &types.EventHLBridgingFee{
		HookId:  hookId.String(),
		Sender:  metadata.Address.String(),
		Fee:     fee.String(),
		TokenId: message.Sender.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return feeCoins, nil
}

// QuoteDispatch returns the required fees for dispatching a message
func (f FeeHookHandler) QuoteDispatch(ctx context.Context, _, hookId hyputil.HexAddress, _ hyputil.StandardHookMetadata, message hyputil.HyperlaneMessage) (sdk.Coins, error) {
	fee, err := f.quoteFee(ctx, hookId, message.Sender, message.Body)
	if err != nil {
		return nil, err
	}
	return sdk.NewCoins(fee), nil
}

// quoteFee calculates the fee required for a specific token transfer
func (f FeeHookHandler) quoteFee(ctx context.Context, hookId hyputil.HexAddress, sender hyputil.HexAddress, body []byte) (sdk.Coin, error) {
	// Get the fee hook configuration
	hook, err := f.k.feeHooks.Get(ctx, hookId.GetInternalId())
	if err != nil {
		return sdk.Coin{}, errorsmod.Wrap(err, "get fee hook")
	}

	// Check if we have a fee configuration for this token (sender is the token ID)
	var assetFee *types.HLAssetFee
	for _, fee := range hook.Fees {
		if fee.TokenID == sender.String() {
			assetFee = &fee
			break
		}
	}

	// If no fee configured for this token, return zero fee
	if assetFee == nil {
		return sdk.Coin{}, nil
	}

	// Get token information from warp keeper
	tokenResp, err := f.k.warpQuery.Token(ctx, &warptypes.QueryTokenRequest{Id: sender.String()})
	if err != nil {
		return sdk.Coin{}, errorsmod.Wrap(err, "get token from warp keeper")
	}

	// Parse warp payload to get transfer amount
	payload, err := warptypes.ParseWarpPayload(body)
	if err != nil {
		return sdk.Coin{}, errorsmod.Wrap(err, "parse warp payload")
	}

	// Calculate fee: payload.Amount * outboundFee
	// outboundFee is a LegacyDec, so we need to multiply and truncate to get integer
	transferAmount := math.NewIntFromBigInt(payload.Amount())
	feeAmount := assetFee.OutboundFee.MulInt(transferAmount).TruncateInt()

	return sdk.NewCoin(tokenResp.Token.OriginDenom, feeAmount), nil
}
