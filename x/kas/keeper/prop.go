package keeper

import (
	"context"

	hyputil "github.com/bcp-innovations/hyperlane-cosmos/util"
	warptypes "github.com/bcp-innovations/hyperlane-cosmos/x/warp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/kas/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"
)

func (k *Keeper) Bootstrap(goCtx context.Context, req *types.MsgBootstrap) (*types.MsgBootstrapResponse, error) {
	if req.Authority != k.authority {
		return nil, gerrc.ErrPermissionDenied
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Checks

	if k.Ready(ctx) {
		// we already finished bootstrap in a prior proposal, and someone has already used the bridge
		// which poses a risk
		supply, err := k.warpQ.BridgedSupply(ctx, &warptypes.QueryBridgedSupplyRequest{
			Id: req.TokenId,
		})
		if err != nil {
			return nil, err
		}
		if !supply.BridgedSupply.Amount.IsZero() {
			panic("tokens already bridged")
		}
	}

	mailbox, err := hyputil.DecodeHexAddress(req.Mailbox)
	if err != nil {
		return nil, err
	}

	ism, err := hyputil.DecodeHexAddress(req.Ism)
	if err != nil {
		return nil, err
	}

	if err := req.Outpoint.ValidateBasic(); err != nil {
		return nil, err
	}

	found, err := k.hypercoreK.MailboxIdExists(ctx, mailbox)
	if err != nil || !found {
		return nil, gerrc.ErrNotFound.Wrap("mailbox")
	}

	if k.hypercoreK.AssertIsmExists(ctx, ism) != nil {
		return nil, gerrc.ErrNotFound.Wrap("ism")
	}

	empty, err := k.WithdrawalsEmpty(ctx)
	if err != nil {
		return nil, err
	}
	if !empty {
		err := gerrc.ErrDataLoss.Wrap("withdrawals not empty: module has already been used, rollback is undefined")
		panic(err)
	}

	// Sets

	if err := k.mailbox.Set(ctx, req.Mailbox); err != nil {
		return nil, err
	}

	if err := k.ism.Set(ctx, req.Ism); err != nil {
		return nil, err
	}

	if err := k.outpoint.Set(ctx, *req.Outpoint); err != nil {
		return nil, err
	}

	if err := k.bootstrapped.Set(ctx, true); err != nil {
		return nil, err
	}

	if err := uevent.EmitTypedEvent(ctx, &types.EventBootstrap{}); err != nil {
		return nil, err
	}

	return &types.MsgBootstrapResponse{}, nil
}
