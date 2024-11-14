package keeper

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"

	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
	rollapptypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
)

// FindMatchingClient returns the client id of the first IBC client which can be set as the canonical client for the given rollapp.
// The canonical client criteria are:
// 1. The client must be a tendermint client.
// 2. The client state must match the expected client params as configured by the module
// 3. All the existing consensus states much match the corresponding height rollapp block descriptors
func (k Keeper) FindMatchingClient(ctx sdk.Context, sInfo *rollapptypes.StateInfo) (clientID string, stateCompatible bool) {
	k.ibcClientKeeper.IterateClientStates(ctx, nil, func(client string, cs exported.ClientState) bool {
		err := k.validClient(ctx, client, cs, sInfo)
		if err == nil {
			clientID = client
			stateCompatible = true
			return true
		}
		if !errorsmod.IsOf(err, errChainIDMismatch) {
			// Log the error with key-value pairs
			ctx.Logger().Debug("tried to validate rollapp against light client for same chain id",
				"rollapp", sInfo.GetRollappId(),
				"client", client,
				"err", err,
			)
		}
		return false
	})
	return
}

func (k Keeper) GetCanonicalClient(ctx sdk.Context, rollappId string) (string, bool) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetRollappClientKey(rollappId))
	if bz == nil {
		return "", false
	}
	return string(bz), true
}

func (k Keeper) SetCanonicalClient(ctx sdk.Context, rollappId string, clientID string) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetRollappClientKey(rollappId), []byte(clientID))
	store.Set(types.CanonicalClientKey(clientID), []byte(rollappId))
	// TODO: event and log
}

func (k Keeper) GetAllCanonicalClients(ctx sdk.Context) (clients []types.CanonicalClient) {
	store := ctx.KVStore(k.storeKey)
	iterator := sdk.KVStorePrefixIterator(store, types.RollappClientKey)
	defer iterator.Close() // nolint: errcheck
	for ; iterator.Valid(); iterator.Next() {
		clients = append(clients, types.CanonicalClient{
			RollappId:   string(iterator.Key()[1:]),
			IbcClientId: string(iterator.Value()),
		})
	}
	return
}

func (k Keeper) expectedClient() ibctm.ClientState {
	return types.DefaultExpectedCanonicalClientParams()
}

var errChainIDMismatch = errors.New("chain id mismatch")

func (k Keeper) validClient(ctx sdk.Context, clientID string, cs exported.ClientState, sInfo *rollapptypes.StateInfo) error {
	maxHeight := sInfo.GetLatestHeight()
	minHeight := sInfo.StartHeight
	rollappId := sInfo.GetRollappId()

	tmClientState, ok := cs.(*ibctm.ClientState)
	if !ok {
		return errors.New("not tm client")
	}
	if tmClientState.ChainId != rollappId {
		return errChainIDMismatch
	}

	expClient := k.expectedClient()
	if err := types.IsCanonicalClientParamsValid(tmClientState, &expClient); err != nil {
		return errorsmod.Wrap(err, "params")
	}

	atLeastOneMatch := false
	csStore := k.ibcClientKeeper.ClientStore(ctx, clientID)
	var err error
	IterateConsensusStateDescending(csStore, func(h exported.Height) bool {
		// skip future heights
		if h.GetRevisionHeight() >= maxHeight {
			return false
		}

		// iterate until we pass the fraud height
		if h.GetRevisionHeight() < minHeight {
			return true // break
		}

		consensusState, ok := k.ibcClientKeeper.GetClientConsensusState(ctx, clientID, h)
		if !ok {
			return false
		}
		tmConsensusState, ok := consensusState.(*ibctm.ConsensusState)
		if !ok {
			return false
		}

		err = k.ValidateUpdatePessimistically(ctx, sInfo, tmConsensusState, h.GetRevisionHeight())
		if err != nil {
			err = errorsmod.Wrapf(err, "validate pessimistic h: %d", h.GetRevisionHeight())
			return true // break
		}

		atLeastOneMatch = true
		return true // break
	})
	// Need to be sure that at least one consensus state agrees with a state update
	// (There are also no disagreeing consensus states. There may be some consensus states
	// for future state updates, which will incur a fraud if they disagree.)
	if !atLeastOneMatch {
		err = errors.Join(errors.New("no consensus state matches"), err)
	}

	if err != nil {
		return errorsmod.Wrapf(err, "testing client %s for rollapp %s", clientID, rollappId)
	}
	return nil
}

func (k Keeper) ValidateUpdatePessimistically(ctx sdk.Context, sInfo *rollapptypes.StateInfo, consState *ibctm.ConsensusState, h uint64) error {
	bd, ok := sInfo.GetBlockDescriptor(h)
	if !ok {
		return errorsmod.Wrapf(gerrc.ErrInternal, "no block descriptor found for height %d", h)
	}
	nextSeq, err := k.SeqK.RealSequencer(ctx, sInfo.NextSequencerForHeight(h))
	if err != nil {
		return errorsmod.Wrap(errors.Join(err, gerrc.ErrInternal), "get sequencer of state info")
	}
	rollappState := types.RollappState{
		BlockDescriptor:    bd,
		NextBlockSequencer: nextSeq,
	}
	return errorsmod.Wrap(types.CheckCompatibility(*consState, rollappState), "check compatibility")
}
