package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	hyputil "github.com/bcp-innovations/hyperlane-cosmos/util"
	postdispatchtypes "github.com/bcp-innovations/hyperlane-cosmos/x/core/02_post_dispatch/types"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/bridgingfee/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
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
func (f FeeHookHandler) PostDispatch(goCtx context.Context, _, hookId hyputil.HexAddress, metadata hyputil.StandardHookMetadata, message hyputil.HyperlaneMessage, maxFee sdk.Coins) (sdk.Coins, error) {
	// Parse warp payload to get transfer amount
	payload, err := warptypes.ParseWarpPayload(message.Body)
	if err != nil {
		return nil, fmt.Errorf("parse warp payload: %w", err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	fee, err := f.QuoteFeeInBase(ctx, hookId, message.Sender, math.NewIntFromBigIntMut(payload.Amount()))
	if err != nil {
		return nil, err
	}
	if fee.IsZero() {
		// Nothing to charge
		return sdk.NewCoins(), nil
	}
	feeCoins := sdk.NewCoins(fee)

	if !maxFee.IsAllGTE(feeCoins) {
		return sdk.NewCoins(), fmt.Errorf("required fee payment exceeds max fee: %v", maxFee)
	}

	// Accumulate fees on the x/bridgingfee account
	err = f.k.bankKeeper.SendCoinsFromAccountToModule(ctx, metadata.Address, types.ModuleName, feeCoins)
	if err != nil {
		return nil, fmt.Errorf("send fee from sender to x/bridgingfee: %w", err)
	}

	err = uevent.EmitTypedEvent(ctx, &types.EventHLBridgingFee{
		HookId:    hookId,
		Payer:     metadata.Address.String(),
		TokenId:   message.Sender,
		Fee:       fee.String(),
		MessageId: message.Id(),
	})
	if err != nil {
		return nil, fmt.Errorf("emit event: %w", err)
	}

	return feeCoins, nil
}

// QuoteDispatch returns the required fees for dispatching a message
func (f FeeHookHandler) QuoteDispatch(goCtx context.Context, _, hookId hyputil.HexAddress, _ hyputil.StandardHookMetadata, message hyputil.HyperlaneMessage) (sdk.Coins, error) {
	// Parse warp payload to get transfer amount
	payload, err := warptypes.ParseWarpPayload(message.Body)
	if err != nil {
		return nil, fmt.Errorf("parse warp payload: %w", err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	fee, err := f.QuoteFeeInBase(ctx, hookId, message.Sender, math.NewIntFromBigIntMut(payload.Amount()))
	if err != nil {
		return nil, err
	}

	return sdk.NewCoins(fee), nil
}

// QuoteFeeInBase calculates the fee in base denom required for a specific token transfer
func (f FeeHookHandler) QuoteFeeInBase(ctx sdk.Context, hookId hyputil.HexAddress, sender hyputil.HexAddress, transferAmt math.Int) (sdk.Coin, error) {
	// Get the fee hook configuration
	hook, err := f.k.feeHooks.Get(ctx, hookId.GetInternalId())
	if err != nil {
		return sdk.Coin{}, fmt.Errorf("get fee hook: %w", err)
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

	// Get original denom of the token
	tokenResp, err := f.k.warpQuery.Token(ctx, &warptypes.QueryTokenRequest{Id: sender.String()})
	if err != nil {
		return sdk.Coin{}, fmt.Errorf("get token from warp keeper: %w", err)
	}

	// fee = transferAmt * outboundFee
	fee := assetFee.OutboundFee.MulInt(transferAmt).TruncateInt()

	feeCoin := sdk.NewCoin(tokenResp.Token.OriginDenom, fee)

	feeBase, err := f.k.txFeesKeeper.CalcCoinInBaseDenom(ctx, feeCoin)
	if err != nil {
		return sdk.Coin{}, fmt.Errorf("calc fee in base denom: %w", err)
	}

	return feeBase, nil
}
