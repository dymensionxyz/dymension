package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	hyputil "github.com/bcp-innovations/hyperlane-cosmos/util"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/bridgingfee/types"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
)

// FeeHookHandler is a Hyperlane post-dispatch hook that charges protocol fees for outbound token transfers.
// This hook calculates and collects fees based on a specified token type and transfer amount before the HL transfer
// is dispatched from the Hub. The hook can be configured with different fee rates for different tokens.
type FeeHookHandler struct {
	k Keeper
}

// NewFeeHookHandler creates a new FeeHookHandler
func NewFeeHookHandler(k Keeper) FeeHookHandler {
	return FeeHookHandler{k: k}
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
	return types.PostDispatchHookDymProtocolFee
}

// PostDispatch collects fees from the sender for bridging tokens
func (f FeeHookHandler) PostDispatch(goCtx context.Context, _, hookId hyputil.HexAddress, metadata hyputil.StandardHookMetadata, message hyputil.HyperlaneMessage, maxFee sdk.Coins) (sdk.Coins, error) {
	// Parse warp payload to get transfer amount
	payload, err := warptypes.ParseWarpPayload(message.Body)
	if err != nil {
		return nil, fmt.Errorf("parse warp payload: %w", err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	fee, err := f.QuoteFee(ctx, hookId, message.Sender, math.NewIntFromBigIntMut(payload.Amount()))
	if err != nil {
		return nil, err
	}
	if fee.IsZero() {
		// Nothing to charge
		return nil, nil
	}

	if !maxFee.IsAllGTE(fee) {
		return nil, fmt.Errorf("required fee payment exceeds max fee: required %v, max %v", fee, maxFee)
	}

	// Accumulate fees on the x/bridgingfee account
	err = f.k.bankKeeper.SendCoinsFromAccountToModule(ctx, metadata.Address, types.ModuleName, fee)
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

	return fee, nil
}

// QuoteDispatch returns the required fees for dispatching a message
func (f FeeHookHandler) QuoteDispatch(goCtx context.Context, _, hookId hyputil.HexAddress, _ hyputil.StandardHookMetadata, message hyputil.HyperlaneMessage) (sdk.Coins, error) {
	// Parse warp payload to get transfer amount
	payload, err := warptypes.ParseWarpPayload(message.Body)
	if err != nil {
		return nil, fmt.Errorf("parse warp payload: %w", err)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	fee, err := f.QuoteFee(ctx, hookId, message.Sender, math.NewIntFromBigIntMut(payload.Amount()))
	if err != nil {
		return nil, fmt.Errorf("quote fee in base: %w", err)
	}

	return fee, nil
}

// QuoteFee calculates the fee for a specific token transfer. `transferAmt` is in `sender.OriginalDenom`.
func (f FeeHookHandler) QuoteFee(ctx sdk.Context, hookId hyputil.HexAddress, sender hyputil.HexAddress, transferAmt math.Int) (sdk.Coins, error) {
	// Get the fee hook configuration
	hook, err := f.k.feeHooks.Get(ctx, hookId.GetInternalId())
	if err != nil {
		return nil, fmt.Errorf("get fee hook: %w", err)
	}

	// Check if we have a fee configuration for this token (sender is the token ID)
	var assetFee *types.HLAssetFee
	for _, fee := range hook.Fees {
		if fee.TokenId.Equal(sender) {
			assetFee = &fee
			break
		}
	}

	// If no fee configured for this token, return zero fee
	if assetFee == nil {
		return nil, nil
	}

	// Get original denom of the token
	tokenResp, err := f.k.warpQuery.Token(ctx, &warptypes.QueryTokenRequest{Id: sender.String()})
	if err != nil {
		return nil, fmt.Errorf("get token from warp keeper: %w", err)
	}

	// fee = transferAmt * outboundFee
	fee := assetFee.OutboundFee.MulInt(transferAmt).TruncateInt()
	return sdk.NewCoins(sdk.NewCoin(tokenResp.Token.OriginDenom, fee)), nil
}
