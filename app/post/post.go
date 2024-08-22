package post

import (
	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	lightclientkeeper "github.com/dymensionxyz/dymension/v3/x/lightclient/keeper"

	lightclientpost "github.com/dymensionxyz/dymension/v3/x/lightclient/post"
)

// HandlerOptions are the options required for constructing a default SDK PostHandler.
type HandlerOptions struct {
	IBCKeeper         *ibckeeper.Keeper
	LightClientKeeper *lightclientkeeper.Keeper
}

func NewPostHandler(options HandlerOptions) (sdk.PostHandler, error) {
	if err := options.validate(); err != nil {
		return nil, err
	}
	postDecorators := []sdk.PostDecorator{
		lightclientpost.NewIBCMessagesDecorator(*options.LightClientKeeper, options.IBCKeeper.ClientKeeper),
	}

	return sdk.ChainPostDecorators(postDecorators...), nil
}

func (options HandlerOptions) validate() error {
	if options.IBCKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "ibc keeper is required for PostHandler")
	}
	if options.LightClientKeeper == nil {
		return errorsmod.Wrap(errortypes.ErrLogic, "light client keeper is required for PostHandler")
	}
	return nil
}
