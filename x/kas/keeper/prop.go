package keeper

import (
	"context"

	hyputil "github.com/bcp-innovations/hyperlane-cosmos/util"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/dymension/v3/x/kas/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func (k *Keeper) Bootstrap(goCtx context.Context, req *types.MsgBootstrap) (*types.MsgBootstrapResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if req.Authority != k.authority {
		return nil, gerrc.ErrPermissionDenied
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

	if err := k.mailbox.Set(ctx, req.Mailbox); err != nil {
		return nil, err
	}

	if err := k.ism.Set(ctx, req.Ism); err != nil {
		return nil, err
	}

	if err := k.outpoint.Set(ctx, *req.Outpoint); err != nil {
		return nil, err
	}

	return &types.MsgBootstrapResponse{}, nil
}
