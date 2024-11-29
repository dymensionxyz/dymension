package keeper

import (
	"errors"

	errorsmod "cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"
	"github.com/dymensionxyz/gerr-cosmos/gerrc"
	"github.com/dymensionxyz/sdk-utils/utils/uevent"

	"github.com/dymensionxyz/dymension/v3/x/lightclient/types"
)

// intended to be called by relayer, but can be called by anyone
// verifies that the suggested client is safe to designate canonical and matches state updates from the sequencer
func (k *Keeper) TrySetCanonicalClient(ctx sdk.Context, clientID string) error {
	clientStateI, ok := k.ibcClientKeeper.GetClientState(ctx, clientID)
	if !ok {
		return gerrc.ErrNotFound.Wrap("client")
	}

	clientState, ok := clientStateI.(*ibctm.ClientState)
	if !ok {
		return gerrc.ErrInvalidArgument.Wrap("not tm client")
	}

	chainID := clientState.ChainId
	_, ok = k.rollappKeeper.GetRollapp(ctx, chainID)
	if !ok {
		return gerrc.ErrNotFound.Wrap("rollapp")
	}
	rollappID := chainID

	_, ok = k.GetCanonicalClient(ctx, rollappID)
	if ok {
		return gerrc.ErrAlreadyExists.Wrap("canonical client for rollapp")
	}

	latestHeight, ok := k.rollappKeeper.GetLatestHeight(ctx, rollappID)
	if !ok {
		return gerrc.ErrNotFound.Wrap("latest rollapp height")
	}

	err := k.validClient(ctx, clientID, clientState, rollappID, latestHeight)
	if err != nil {
		return errorsmod.Wrap(err, "unsafe to mark client canonical: check that sequencer has posted a recent state update")
	}

	k.SetCanonicalClient(ctx, rollappID, clientID)

	if err := uevent.EmitTypedEvent(ctx, &types.EventSetCanonicalClient{
		RollappId: rollappID,
		ClientId:  clientID,
	}); err != nil {
		return errorsmod.Wrap(err, "emit typed event")
	}

	return nil
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

var (
	ErrNoMatch        = gerrc.ErrFailedPrecondition.Wrap("not at least one cons state matches the rollapp state")
	ErrMismatch       = gerrc.ErrInvalidArgument.Wrap("consensus state mismatch")
	ErrParamsMismatch = gerrc.ErrInvalidArgument.Wrap("params")
)

// The canonical client criteria are:
// 1. The client must be a tendermint client.
// 2. The client state must match the expected client params as configured by the module
// 3. All the existing consensus states much match the corresponding height rollapp block descriptors
func (k Keeper) validClient(ctx sdk.Context, clientID string, cs *ibctm.ClientState, rollappId string, maxHeight uint64) error {
	log := k.Logger(ctx).With("component", "valid client func", "rollapp", rollappId, "client", clientID)

	log.Debug("top of func", "max height", maxHeight, "gas", ctx.GasMeter().GasConsumed())

	expClient := k.expectedClient()

	if err := types.IsCanonicalClientParamsValid(cs, &expClient); err != nil {
		return errors.Join(err, ErrParamsMismatch)
	}

	// FIXME: No need to get all consensus states. should iterate over the consensus states
	res, err := k.ibcClientKeeper.ConsensusStateHeights(ctx, &ibcclienttypes.QueryConsensusStateHeightsRequest{
		ClientId:   clientID,
		Pagination: &query.PageRequest{Limit: maxHeight},
	})
	log.Debug("after fetch heights", "max height", maxHeight, "gas", ctx.GasMeter().GasConsumed())
	if err != nil {
		return errorsmod.Wrap(err, "cons state heights")
	}
	atLeastOneMatch := false
	for _, consensusHeight := range res.ConsensusStateHeights {
		log.Debug("after fetch heights", "cons state height", consensusHeight.RevisionHeight, "gas", ctx.GasMeter().GasConsumed())
		h := consensusHeight.GetRevisionHeight()
		if maxHeight <= h {
			break
		}
		consensusState, _ := k.ibcClientKeeper.GetClientConsensusState(ctx, clientID, consensusHeight)
		tmConsensusState, _ := consensusState.(*ibctm.ConsensusState)
		stateInfoH, err := k.rollappKeeper.FindStateInfoByHeight(ctx, rollappId, h)
		if err != nil {
			return errorsmod.Wrapf(err, "find state info by height h: %d", h)
		}
		stateInfoHplus1, err := k.rollappKeeper.FindStateInfoByHeight(ctx, rollappId, h+1)
		if err != nil {
			return errorsmod.Wrapf(err, "find state info by height h+1: %d", h+1)
		}
		bd, _ := stateInfoH.GetBlockDescriptor(h)

		nextSeq, err := k.SeqK.RealSequencer(ctx, stateInfoHplus1.Sequencer)
		if err != nil {
			return errorsmod.Wrap(err, "get sequencer")
		}
		rollappState := types.RollappState{
			BlockDescriptor:    bd,
			NextBlockSequencer: nextSeq,
		}
		err = types.CheckCompatibility(*tmConsensusState, rollappState)
		if err != nil {
			return errorsmod.Wrapf(errors.Join(ErrMismatch, err), "check compatibility: height: %d", h)
		}
		atLeastOneMatch = true
	}
	// Need to be sure that at least one consensus state agrees with a state update
	// (There are also no disagreeing consensus states. There may be some consensus states
	// for future state updates, which will incur a fraud if they disagree.)
	if !atLeastOneMatch {
		return ErrNoMatch
	}
	return nil
}
