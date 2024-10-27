package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (hook rollappHook) OnHardFork(sdk.Context, string, uint64) error {
	/*
		// freeze IBC client state
		clientState, ok := k.ibcClientKeeper.GetClientState(ctx, clientId)
		if !ok {
			return errorsmod.Wrapf(types.ErrInvalidClientState, "client state for clientID %s not found", clientId)
		}

		tmClientState, ok := clientState.(*cometbfttypes.ClientState)
		if !ok {
			return errorsmod.Wrapf(types.ErrInvalidClientState, "client state with ID %s is not a tendermint client state", clientId)
		}

		tmClientState.FrozenHeight = clienttypes.NewHeight(tmClientState.GetLatestHeight().GetRevisionHeight(), tmClientState.GetLatestHeight().GetRevisionNumber())
		k.ibcClientKeeper.SetClientState(ctx, clientId, tmClientState)

	*/
	return nil
}
