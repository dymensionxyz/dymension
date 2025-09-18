package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

func (k msgServer) UpdateParams(goCtx context.Context, msg *dymnstypes.MsgUpdateParams) (*dymnstypes.MsgUpdateParamsResponse, error) {
	err := msg.ValidateBasic()
	if err != nil {
		return nil, err
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Authority != k.authority {
		return nil, errorsmod.Wrap(gerrc.ErrUnauthenticated, "only the gov module can update params")
	}

	moduleParams := k.GetParams(ctx)

	if msg.NewPriceParams != nil {
		// price denom must be the same as txfees base denom
		baseDenom, err := k.txFeesKeeper.GetBaseDenom(ctx)
		if err != nil {
			return nil, err
		}
		if msg.NewPriceParams.PriceDenom != baseDenom {
			return nil, errorsmod.Wrap(gerrc.ErrInvalidArgument, "price denom must be the same as txfees base denom")
		}

		moduleParams.Price = *msg.NewPriceParams
	}

	if msg.NewChainsParams != nil {
		moduleParams.Chains = *msg.NewChainsParams
	}

	if msg.NewMiscParams != nil {
		moduleParams.Misc = *msg.NewMiscParams
	}

	err = k.SetParams(ctx, moduleParams)
	if err != nil {
		return nil, err
	}

	return &dymnstypes.MsgUpdateParamsResponse{}, nil
}
